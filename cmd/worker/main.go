package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"smtp-relay/internal/database"
	"smtp-relay/internal/queue"
	"smtp-relay/internal/worker"
)

func main() {
	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// 加载配置
	if err := loadConfig(); err != nil {
		logger.WithError(err).Fatal("加载配置失败")
	}

	// 设置日志级别
	if level, err := logrus.ParseLevel(viper.GetString("LOG_LEVEL")); err == nil {
		logger.SetLevel(level)
	}

	logger.Info("启动SMTP中继工作进程")

	// 连接MongoDB
	db, err := database.NewMongoDB(
		viper.GetString("MONGODB_URI"),
		viper.GetString("MONGODB_DATABASE"),
		logger,
	)
	if err != nil {
		logger.WithError(err).Fatal("连接MongoDB失败")
	}
	defer db.Close()

	// 创建索引
	if err := db.CreateIndexes(); err != nil {
		logger.WithError(err).Fatal("创建数据库索引失败")
	}

	// 初始化队列服务
	queueConfig := &queue.Config{
		URL:          viper.GetString("RABBITMQ_URL"),
		ExchangeName: viper.GetString("RABBITMQ_EXCHANGE"),
		QueueName:    viper.GetString("RABBITMQ_QUEUE"),
		RoutingKey:   "mail",
	}

	queueService, err := queue.NewService(queueConfig, db, logger)
	if err != nil {
		logger.WithError(err).Fatal("初始化队列服务失败")
	}
	defer queueService.Close()

	// 创建邮件处理器
	processor := worker.NewProcessor(db, logger, queueService)

	// 启动处理器
	processorConfig := &worker.Config{
		WorkerCount:    viper.GetInt("WORKER_COUNT"),
		ProcessTimeout: viper.GetDuration("PROCESS_TIMEOUT"),
		RetryInterval:  viper.GetDuration("RETRY_INTERVAL"),
	}

	if err := processor.Start(processorConfig); err != nil {
		logger.WithError(err).Fatal("启动邮件处理器失败")
	}

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("SMTP中继工作进程启动完成")

	// 等待信号
	<-sigChan
	logger.Info("收到停止信号，正在关闭服务...")

	// 停止处理器
	processor.Stop()

	// 给处理器一些时间来完成正在处理的邮件
	time.Sleep(10 * time.Second)

	logger.Info("SMTP中继工作进程已停止")
}

// loadConfig 加载配置
func loadConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("MONGODB_DATABASE", "smtp_relay")
	viper.SetDefault("RABBITMQ_EXCHANGE", "smtp_relay")
	viper.SetDefault("RABBITMQ_QUEUE", "mail_queue")
	viper.SetDefault("WORKER_COUNT", 5)
	viper.SetDefault("PROCESS_TIMEOUT", "30s")
	viper.SetDefault("RETRY_INTERVAL", "1m")

	// 从环境变量读取
	viper.AutomaticEnv()

	// 尝试读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return nil
}
