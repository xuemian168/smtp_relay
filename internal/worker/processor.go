package worker

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"smtp-relay/internal/database"
	"smtp-relay/internal/models"
	"smtp-relay/internal/queue"
)

// Processor 邮件处理器
type Processor struct {
	db           *database.MongoDB
	logger       *logrus.Logger
	queueService *queue.Service
	smtpConfigs  []*models.SMTPConfig
	stopChan     chan struct{}
}

// Config 处理器配置
type Config struct {
	WorkerCount    int
	ProcessTimeout time.Duration
	RetryInterval  time.Duration
}

// NewProcessor 创建邮件处理器
func NewProcessor(db *database.MongoDB, logger *logrus.Logger, queueService *queue.Service) *Processor {
	return &Processor{
		db:           db,
		logger:       logger,
		queueService: queueService,
		stopChan:     make(chan struct{}),
	}
}

// Start 启动邮件处理器
func (p *Processor) Start(config *Config) error {
	p.logger.Info("启动邮件处理器")

	// 加载SMTP配置
	if err := p.loadSMTPConfigs(); err != nil {
		return fmt.Errorf("加载SMTP配置失败: %w", err)
	}

	// 启动多个工作协程
	for i := 0; i < config.WorkerCount; i++ {
		go p.worker(i, config)
	}

	// 启动配置刷新协程
	go p.configRefresher()

	p.logger.WithField("worker_count", config.WorkerCount).Info("邮件处理器启动完成")
	return nil
}

// Stop 停止邮件处理器
func (p *Processor) Stop() {
	p.logger.Info("停止邮件处理器")
	close(p.stopChan)
}

// worker 工作协程
func (p *Processor) worker(workerID int, config *Config) {
	logger := p.logger.WithField("worker_id", workerID)
	logger.Info("工作协程启动")

	defer func() {
		if r := recover(); r != nil {
			logger.WithField("panic", r).Error("工作协程发生panic")
		}
		logger.Info("工作协程退出")
	}()

	// 消费队列消息
	err := p.queueService.ConsumeMessages(func(message *queue.MailMessage) error {
		return p.processMessage(message, logger)
	})

	if err != nil {
		logger.WithError(err).Error("消费队列消息失败")
	}
}

// processMessage 处理邮件消息
func (p *Processor) processMessage(message *queue.MailMessage, logger *logrus.Entry) error {
	logger = logger.WithFields(logrus.Fields{
		"mail_id":    message.MailLogID.Hex(),
		"message_id": message.MailLogID.Hex(),
		"from":       message.From,
		"to_count":   len(message.To),
	})

	logger.Info("开始处理邮件")

	// 更新邮件状态为发送中
	if err := p.updateMailStatus(message.MailLogID, "sending", "", 0); err != nil {
		logger.WithError(err).Error("更新邮件状态失败")
		return err
	}

	// 选择SMTP服务器
	smtpConfig := p.selectSMTPServer()
	if smtpConfig == nil {
		err := fmt.Errorf("没有可用的SMTP服务器")
		logger.Error(err.Error())
		p.updateMailStatus(message.MailLogID, "failed", err.Error(), 0)
		return err
	}

	logger.WithFields(logrus.Fields{
		"smtp_host": smtpConfig.Host,
		"smtp_port": smtpConfig.Port,
	}).Info("选择SMTP服务器")

	// 发送邮件
	attempts := 0
	var lastError error

	for attempts < 3 {
		attempts++
		logger.WithField("attempt", attempts).Info("尝试发送邮件")

		// 更新尝试次数
		if err := p.updateMailStatus(message.MailLogID, "sending", "", attempts); err != nil {
			logger.WithError(err).Warn("更新尝试次数失败")
		}

		// 发送邮件
		if err := p.sendMail(smtpConfig, message, logger); err != nil {
			lastError = err
			logger.WithError(err).WithField("attempt", attempts).Warn("发送邮件失败")

			// 如果是临时错误，等待后重试
			if p.isTemporaryError(err) && attempts < 3 {
				delay := time.Duration(attempts) * 30 * time.Second
				logger.WithField("delay", delay).Info("等待后重试")
				time.Sleep(delay)
				continue
			}
		} else {
			// 发送成功
			logger.Info("邮件发送成功")
			now := time.Now()
			if err := p.updateMailStatusWithCompletion(message.MailLogID, "sent", "", attempts, &now); err != nil {
				logger.WithError(err).Error("更新邮件完成状态失败")
			}
			return nil
		}
	}

	// 所有尝试都失败
	logger.WithError(lastError).Error("邮件发送失败，已达最大重试次数")
	if err := p.updateMailStatus(message.MailLogID, "failed", lastError.Error(), attempts); err != nil {
		logger.WithError(err).Error("更新邮件失败状态失败")
	}

	return lastError
}

// sendMail 发送邮件
func (p *Processor) sendMail(config *models.SMTPConfig, message *queue.MailMessage, logger *logrus.Entry) error {
	// 建立SMTP连接
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	var client *smtp.Client
	var err error

	if config.TLS {
		// 使用TLS连接
		client, err = p.connectWithTLS(addr, config)
	} else {
		// 使用普通连接
		client, err = smtp.Dial(addr)
	}

	if err != nil {
		return fmt.Errorf("连接SMTP服务器失败: %w", err)
	}
	defer client.Close()

	// SMTP认证
	if config.Username != "" && config.Password != "" {
		auth := smtp.PlainAuth("", config.Username, config.Password, config.Host)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP认证失败: %w", err)
		}
	}

	// 设置发件人
	if err := client.Mail(message.From); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	for _, to := range message.To {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败 (%s): %w", to, err)
		}
	}

	// 发送邮件数据
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始数据传输失败: %w", err)
	}

	// 构建邮件头
	headers := p.buildMailHeaders(message)

	// 写入邮件头
	if _, err := writer.Write([]byte(headers)); err != nil {
		writer.Close()
		return fmt.Errorf("写入邮件头失败: %w", err)
	}

	// 写入邮件体
	if _, err := writer.Write(message.Body); err != nil {
		writer.Close()
		return fmt.Errorf("写入邮件体失败: %w", err)
	}

	// 完成数据传输
	if err := writer.Close(); err != nil {
		return fmt.Errorf("完成数据传输失败: %w", err)
	}

	// 退出SMTP会话
	if err := client.Quit(); err != nil {
		logger.WithError(err).Warn("退出SMTP会话失败")
	}

	return nil
}

// connectWithTLS 使用TLS连接SMTP服务器
func (p *Processor) connectWithTLS(addr string, config *models.SMTPConfig) (*smtp.Client, error) {
	// 这里简化实现，实际应该根据端口和配置选择合适的TLS连接方式
	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, err
	}

	// 如果服务器支持STARTTLS，则启用
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(nil); err != nil {
			client.Close()
			return nil, fmt.Errorf("启用STARTTLS失败: %w", err)
		}
	}

	return client, nil
}

// buildMailHeaders 构建邮件头
func (p *Processor) buildMailHeaders(message *queue.MailMessage) string {
	headers := strings.Builder{}

	// 基本头部
	headers.WriteString(fmt.Sprintf("From: %s\r\n", message.From))
	headers.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(message.To, ", ")))
	headers.WriteString(fmt.Sprintf("Subject: %s\r\n", message.Subject))
	headers.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	headers.WriteString("MIME-Version: 1.0\r\n")
	headers.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	headers.WriteString("Content-Transfer-Encoding: 8bit\r\n")

	// 添加邮件ID
	headers.WriteString(fmt.Sprintf("Message-ID: <%s>\r\n", message.MailLogID.Hex()))

	// 添加中继信息
	headers.WriteString("X-Mailer: SMTP-Relay-Service\r\n")
	headers.WriteString(fmt.Sprintf("X-Relay-Time: %s\r\n", time.Now().Format(time.RFC3339)))

	// 空行分隔头部和正文
	headers.WriteString("\r\n")

	return headers.String()
}

// selectSMTPServer 选择SMTP服务器
func (p *Processor) selectSMTPServer() *models.SMTPConfig {
	// 简单的轮询选择，实际应该根据负载、成功率等因素选择
	for _, config := range p.smtpConfigs {
		if config.Active {
			return config
		}
	}
	return nil
}

// isTemporaryError 判断是否为临时错误
func (p *Processor) isTemporaryError(err error) bool {
	errStr := err.Error()

	// 常见的临时错误
	temporaryErrors := []string{
		"connection refused",
		"connection timeout",
		"temporary failure",
		"try again later",
		"service unavailable",
		"too many connections",
		"rate limit",
		"4.", // 4xx SMTP错误码通常是临时错误
	}

	for _, tempErr := range temporaryErrors {
		if strings.Contains(strings.ToLower(errStr), tempErr) {
			return true
		}
	}

	return false
}

// updateMailStatus 更新邮件状态
func (p *Processor) updateMailStatus(mailLogID primitive.ObjectID, status, errorMessage string, attempts int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := p.db.GetCollection("mail_logs")

	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"attempts":     attempts,
			"last_attempt": time.Now(),
		},
	}

	if errorMessage != "" {
		update["$set"].(bson.M)["error_message"] = errorMessage
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": mailLogID}, update)
	return err
}

// updateMailStatusWithCompletion 更新邮件状态并设置完成时间
func (p *Processor) updateMailStatusWithCompletion(mailLogID primitive.ObjectID, status, errorMessage string, attempts int, completedAt *time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := p.db.GetCollection("mail_logs")

	update := bson.M{
		"$set": bson.M{
			"status":       status,
			"attempts":     attempts,
			"last_attempt": time.Now(),
		},
	}

	if errorMessage != "" {
		update["$set"].(bson.M)["error_message"] = errorMessage
	}

	if completedAt != nil {
		update["$set"].(bson.M)["completed_at"] = *completedAt
	}

	_, err := collection.UpdateOne(ctx, bson.M{"_id": mailLogID}, update)
	return err
}

// loadSMTPConfigs 加载SMTP配置
func (p *Processor) loadSMTPConfigs() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := p.db.GetCollection("smtp_configs")
	cursor, err := collection.Find(ctx, bson.M{"active": true})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var configs []*models.SMTPConfig
	if err := cursor.All(ctx, &configs); err != nil {
		return err
	}

	p.smtpConfigs = configs
	p.logger.WithField("config_count", len(configs)).Info("加载SMTP配置完成")

	return nil
}

// configRefresher 配置刷新协程
func (p *Processor) configRefresher() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.loadSMTPConfigs(); err != nil {
				p.logger.WithError(err).Error("刷新SMTP配置失败")
			}
		case <-p.stopChan:
			return
		}
	}
}

// GetStats 获取处理器统计信息
func (p *Processor) GetStats() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := p.db.GetCollection("mail_logs")

	// 统计各种状态的邮件数量
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$status",
				"count": bson.M{"$sum": 1},
			},
		},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	statusCounts := make(map[string]int)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		statusCounts[result.ID] = result.Count
	}

	// 获取队列统计
	queueStats, err := p.queueService.GetQueueStats()
	if err != nil {
		p.logger.WithError(err).Warn("获取队列统计失败")
		queueStats = make(map[string]interface{})
	}

	stats := map[string]interface{}{
		"mail_status_counts": statusCounts,
		"queue_stats":        queueStats,
		"smtp_configs":       len(p.smtpConfigs),
	}

	return stats, nil
}
