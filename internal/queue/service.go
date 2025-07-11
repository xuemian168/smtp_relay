package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"smtp-relay/internal/database"
	"smtp-relay/internal/models"
)

// Service RabbitMQ队列服务
type Service struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	db           *database.MongoDB
	logger       *logrus.Logger
	exchangeName string
	queueName    string
	routingKey   string
}

// Config 队列服务配置
type Config struct {
	URL          string
	ExchangeName string
	QueueName    string
	RoutingKey   string
}

// MailMessage 邮件消息结构
type MailMessage struct {
	MailLogID primitive.ObjectID `json:"mail_log_id"`
	From      string             `json:"from"`
	To        []string           `json:"to"`
	Subject   string             `json:"subject"`
	Body      []byte             `json:"body"`
	Priority  int                `json:"priority"` // 0-9, 9为最高优先级
	CreatedAt time.Time          `json:"created_at"`
}

// NewService 创建队列服务
func NewService(config *Config, db *database.MongoDB, logger *logrus.Logger) (*Service, error) {
	// 连接RabbitMQ
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		logger.WithError(err).Error("连接RabbitMQ失败")
		return nil, err
	}

	// 创建通道
	channel, err := conn.Channel()
	if err != nil {
		logger.WithError(err).Error("创建RabbitMQ通道失败")
		conn.Close()
		return nil, err
	}

	service := &Service{
		conn:         conn,
		channel:      channel,
		db:           db,
		logger:       logger,
		exchangeName: config.ExchangeName,
		queueName:    config.QueueName,
		routingKey:   config.RoutingKey,
	}

	// 初始化队列和交换器
	if err := service.setupQueue(); err != nil {
		logger.WithError(err).Error("初始化队列失败")
		service.Close()
		return nil, err
	}

	logger.Info("RabbitMQ队列服务初始化成功")
	return service, nil
}

// setupQueue 设置队列和交换器
func (s *Service) setupQueue() error {
	// 声明交换器
	err := s.channel.ExchangeDeclare(
		s.exchangeName, // 交换器名称
		"direct",       // 交换器类型
		true,           // 持久化
		false,          // 自动删除
		false,          // 内部使用
		false,          // 不等待
		nil,            // 参数
	)
	if err != nil {
		return fmt.Errorf("声明交换器失败: %w", err)
	}

	// 声明主队列
	queue, err := s.channel.QueueDeclare(
		s.queueName, // 队列名称
		true,        // 持久化
		false,       // 自动删除
		false,       // 独占
		false,       // 不等待
		amqp.Table{
			"x-message-ttl":             int32(24 * 60 * 60 * 1000), // 24小时TTL
			"x-dead-letter-exchange":    s.exchangeName + ".dlx",    // 死信交换器
			"x-dead-letter-routing-key": "failed",                   // 死信路由键
		},
	)
	if err != nil {
		return fmt.Errorf("声明队列失败: %w", err)
	}

	// 绑定队列到交换器
	err = s.channel.QueueBind(
		queue.Name,     // 队列名称
		s.routingKey,   // 路由键
		s.exchangeName, // 交换器名称
		false,          // 不等待
		nil,            // 参数
	)
	if err != nil {
		return fmt.Errorf("绑定队列失败: %w", err)
	}

	// 创建死信交换器和队列
	if err := s.setupDeadLetterQueue(); err != nil {
		return fmt.Errorf("创建死信队列失败: %w", err)
	}

	// 创建延迟队列
	if err := s.setupDelayQueue(); err != nil {
		return fmt.Errorf("创建延迟队列失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"exchange": s.exchangeName,
		"queue":    s.queueName,
	}).Info("队列设置完成")

	return nil
}

// setupDeadLetterQueue 设置死信队列
func (s *Service) setupDeadLetterQueue() error {
	dlxName := s.exchangeName + ".dlx"
	dlqName := s.queueName + ".failed"

	// 声明死信交换器
	err := s.channel.ExchangeDeclare(
		dlxName,  // 交换器名称
		"direct", // 交换器类型
		true,     // 持久化
		false,    // 自动删除
		false,    // 内部使用
		false,    // 不等待
		nil,      // 参数
	)
	if err != nil {
		return err
	}

	// 声明死信队列
	dlq, err := s.channel.QueueDeclare(
		dlqName, // 队列名称
		true,    // 持久化
		false,   // 自动删除
		false,   // 独占
		false,   // 不等待
		nil,     // 参数
	)
	if err != nil {
		return err
	}

	// 绑定死信队列
	err = s.channel.QueueBind(
		dlq.Name, // 队列名称
		"failed", // 路由键
		dlxName,  // 交换器名称
		false,    // 不等待
		nil,      // 参数
	)
	if err != nil {
		return err
	}

	return nil
}

// setupDelayQueue 设置延迟队列
func (s *Service) setupDelayQueue() error {
	delayQueueName := s.queueName + ".delay"

	// 声明延迟队列
	_, err := s.channel.QueueDeclare(
		delayQueueName, // 队列名称
		true,           // 持久化
		false,          // 自动删除
		false,          // 独占
		false,          // 不等待
		amqp.Table{
			"x-dead-letter-exchange":    s.exchangeName, // 死信交换器指向主交换器
			"x-dead-letter-routing-key": s.routingKey,   // 死信路由键
		},
	)
	if err != nil {
		return err
	}

	return nil
}

// EnqueueMail 将邮件加入队列
func (s *Service) EnqueueMail(mailLog *models.MailLog, body []byte) error {
	// 保存MailLog到数据库
	collection := s.db.GetCollection("mail_logs")
	result, err := collection.InsertOne(context.Background(), mailLog)
	if err != nil {
		s.logger.WithError(err).Error("保存MailLog失败")
		return err
	}

	mailLog.ID = result.InsertedID.(primitive.ObjectID)

	// 创建队列消息
	message := &MailMessage{
		MailLogID: mailLog.ID,
		From:      mailLog.From,
		To:        mailLog.To,
		Subject:   mailLog.Subject,
		Body:      body,
		Priority:  s.calculatePriority(mailLog),
		CreatedAt: mailLog.CreatedAt,
	}

	// 序列化消息
	messageBody, err := json.Marshal(message)
	if err != nil {
		s.logger.WithError(err).Error("序列化邮件消息失败")
		return err
	}

	// 发布消息到队列
	err = s.channel.Publish(
		s.exchangeName, // 交换器
		s.routingKey,   // 路由键
		false,          // 强制
		false,          // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBody,
			DeliveryMode: amqp.Persistent, // 持久化消息
			Priority:     uint8(message.Priority),
			Timestamp:    time.Now(),
			MessageId:    mailLog.MessageID,
		},
	)

	if err != nil {
		s.logger.WithError(err).Error("发布邮件消息失败")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"message_id": mailLog.MessageID,
		"mail_id":    mailLog.ID.Hex(),
		"priority":   message.Priority,
	}).Info("邮件已加入队列")

	return nil
}

// EnqueueDelayedMail 将邮件加入延迟队列
func (s *Service) EnqueueDelayedMail(mailLog *models.MailLog, body []byte, delay time.Duration) error {
	delayQueueName := s.queueName + ".delay"

	// 创建队列消息
	message := &MailMessage{
		MailLogID: mailLog.ID,
		From:      mailLog.From,
		To:        mailLog.To,
		Subject:   mailLog.Subject,
		Body:      body,
		Priority:  s.calculatePriority(mailLog),
		CreatedAt: mailLog.CreatedAt,
	}

	// 序列化消息
	messageBody, err := json.Marshal(message)
	if err != nil {
		s.logger.WithError(err).Error("序列化延迟邮件消息失败")
		return err
	}

	// 发布消息到延迟队列
	err = s.channel.Publish(
		"",             // 使用默认交换器
		delayQueueName, // 路由到延迟队列
		false,          // 强制
		false,          // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBody,
			DeliveryMode: amqp.Persistent, // 持久化消息
			Priority:     uint8(message.Priority),
			Timestamp:    time.Now(),
			MessageId:    mailLog.MessageID,
			Expiration:   fmt.Sprintf("%d", delay.Milliseconds()), // 设置过期时间
		},
	)

	if err != nil {
		s.logger.WithError(err).Error("发布延迟邮件消息失败")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"message_id": mailLog.MessageID,
		"mail_id":    mailLog.ID.Hex(),
		"delay":      delay.String(),
	}).Info("邮件已加入延迟队列")

	return nil
}

// ConsumeMessages 消费队列消息
func (s *Service) ConsumeMessages(handler func(*MailMessage) error) error {
	// 设置QoS
	err := s.channel.Qos(
		10,    // 预取数量
		0,     // 预取大小
		false, // 全局设置
	)
	if err != nil {
		return fmt.Errorf("设置QoS失败: %w", err)
	}

	// 开始消费消息
	messages, err := s.channel.Consume(
		s.queueName, // 队列名称
		"",          // 消费者标签
		false,       // 自动确认
		false,       // 独占
		false,       // 不等待
		false,       // 参数
		nil,         // 参数
	)
	if err != nil {
		return fmt.Errorf("开始消费消息失败: %w", err)
	}

	s.logger.Info("开始消费队列消息")

	// 处理消息
	for delivery := range messages {
		var message MailMessage
		if err := json.Unmarshal(delivery.Body, &message); err != nil {
			s.logger.WithError(err).Error("反序列化消息失败")
			delivery.Nack(false, false) // 拒绝消息，不重新入队
			continue
		}

		s.logger.WithFields(logrus.Fields{
			"message_id": delivery.MessageId,
			"mail_id":    message.MailLogID.Hex(),
		}).Info("处理邮件消息")

		// 调用处理函数
		if err := handler(&message); err != nil {
			s.logger.WithError(err).Error("处理邮件失败")

			// 检查重试次数
			retryCount := s.getRetryCount(delivery.Headers)
			if retryCount < 3 {
				// 重新入队延迟处理
				delay := s.calculateRetryDelay(retryCount)
				if err := s.requeueWithDelay(&message, delay, retryCount+1); err != nil {
					s.logger.WithError(err).Error("重新入队失败")
				}
			} else {
				s.logger.WithField("mail_id", message.MailLogID.Hex()).Warn("邮件重试次数已达上限，发送到死信队列")
			}

			delivery.Nack(false, false) // 拒绝消息
		} else {
			delivery.Ack(false) // 确认消息
		}
	}

	return nil
}

// calculatePriority 计算邮件优先级
func (s *Service) calculatePriority(mailLog *models.MailLog) int {
	priority := 5 // 默认优先级

	// 根据邮件大小调整优先级
	if mailLog.Size < 1024*1024 { // 小于1MB
		priority += 1
	} else if mailLog.Size > 10*1024*1024 { // 大于10MB
		priority -= 1
	}

	// 根据收件人数量调整优先级
	if len(mailLog.To) == 1 {
		priority += 1
	} else if len(mailLog.To) > 10 {
		priority -= 1
	}

	// 确保优先级在0-9范围内
	if priority < 0 {
		priority = 0
	} else if priority > 9 {
		priority = 9
	}

	return priority
}

// getRetryCount 获取重试次数
func (s *Service) getRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}

	if count, ok := headers["retry-count"]; ok {
		if retryCount, ok := count.(int32); ok {
			return int(retryCount)
		}
	}

	return 0
}

// calculateRetryDelay 计算重试延迟
func (s *Service) calculateRetryDelay(retryCount int) time.Duration {
	// 指数退避算法
	delays := []time.Duration{
		1 * time.Minute,  // 第一次重试：1分钟
		5 * time.Minute,  // 第二次重试：5分钟
		15 * time.Minute, // 第三次重试：15分钟
	}

	if retryCount < len(delays) {
		return delays[retryCount]
	}

	return 30 * time.Minute // 默认30分钟
}

// requeueWithDelay 重新入队延迟处理
func (s *Service) requeueWithDelay(message *MailMessage, delay time.Duration, retryCount int) error {
	// 更新消息的重试信息
	messageBody, err := json.Marshal(message)
	if err != nil {
		return err
	}

	delayQueueName := s.queueName + ".delay"

	// 发布到延迟队列
	err = s.channel.Publish(
		"",             // 使用默认交换器
		delayQueueName, // 路由到延迟队列
		false,          // 强制
		false,          // 立即
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         messageBody,
			DeliveryMode: amqp.Persistent, // 持久化消息
			Priority:     uint8(message.Priority),
			Timestamp:    time.Now(),
			MessageId:    message.MailLogID.Hex(),
			Expiration:   fmt.Sprintf("%d", delay.Milliseconds()),
			Headers: amqp.Table{
				"retry-count": int32(retryCount),
			},
		},
	)

	if err != nil {
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"mail_id":     message.MailLogID.Hex(),
		"retry_count": retryCount,
		"delay":       delay.String(),
	}).Info("邮件已重新入队延迟处理")

	return nil
}

// GetQueueStats 获取队列统计信息
func (s *Service) GetQueueStats() (map[string]interface{}, error) {
	queue, err := s.channel.QueueInspect(s.queueName)
	if err != nil {
		return nil, err
	}

	dlqName := s.queueName + ".failed"
	dlq, err := s.channel.QueueInspect(dlqName)
	if err != nil {
		s.logger.WithError(err).Warn("获取死信队列统计失败")
		dlq = amqp.Queue{} // 使用空队列信息
	}

	delayQueueName := s.queueName + ".delay"
	delayQueue, err := s.channel.QueueInspect(delayQueueName)
	if err != nil {
		s.logger.WithError(err).Warn("获取延迟队列统计失败")
		delayQueue = amqp.Queue{} // 使用空队列信息
	}

	stats := map[string]interface{}{
		"main_queue": map[string]interface{}{
			"name":      queue.Name,
			"messages":  queue.Messages,
			"consumers": queue.Consumers,
		},
		"dead_letter_queue": map[string]interface{}{
			"name":      dlq.Name,
			"messages":  dlq.Messages,
			"consumers": dlq.Consumers,
		},
		"delay_queue": map[string]interface{}{
			"name":      delayQueue.Name,
			"messages":  delayQueue.Messages,
			"consumers": delayQueue.Consumers,
		},
	}

	return stats, nil
}

// Close 关闭队列服务
func (s *Service) Close() error {
	if s.channel != nil {
		if err := s.channel.Close(); err != nil {
			s.logger.WithError(err).Error("关闭RabbitMQ通道失败")
		}
	}

	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			s.logger.WithError(err).Error("关闭RabbitMQ连接失败")
		}
	}

	s.logger.Info("RabbitMQ队列服务已关闭")
	return nil
}

// IsConnected 检查连接状态
func (s *Service) IsConnected() bool {
	return s.conn != nil && !s.conn.IsClosed()
}

// Reconnect 重新连接
func (s *Service) Reconnect(config *Config) error {
	s.logger.Info("尝试重新连接RabbitMQ")

	// 关闭现有连接
	s.Close()

	// 重新连接
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		s.logger.WithError(err).Error("重新连接RabbitMQ失败")
		return err
	}

	// 创建新通道
	channel, err := conn.Channel()
	if err != nil {
		s.logger.WithError(err).Error("创建新RabbitMQ通道失败")
		conn.Close()
		return err
	}

	s.conn = conn
	s.channel = channel

	// 重新设置队列
	if err := s.setupQueue(); err != nil {
		s.logger.WithError(err).Error("重新设置队列失败")
		s.Close()
		return err
	}

	s.logger.Info("RabbitMQ重新连接成功")
	return nil
}
