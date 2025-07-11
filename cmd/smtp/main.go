package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"smtp-relay/internal/auth"
	"smtp-relay/internal/database"
	"smtp-relay/internal/queue"
	"smtp-relay/internal/services"
	"smtp-relay/internal/smtp"
)

func main() {
	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// 从环境变量读取配置
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017/smtp_relay")
	mongoDatabase := getEnv("MONGODB_DATABASE", "smtp_relay")
	rabbitmqURL := getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/")
	rabbitmqExchange := getEnv("RABBITMQ_EXCHANGE", "smtp_relay")
	rabbitmqQueue := getEnv("RABBITMQ_QUEUE", "mail_queue")
	smtpHost := getEnv("SMTP_HOST", "0.0.0.0")
	smtpDomain := getEnv("SMTP_DOMAIN", "localhost")
	secretKey := getEnv("API_SECRET_KEY", "your-secret-key-change-in-production")

	// 设置日志级别
	if logLevel := getEnv("LOG_LEVEL", "info"); logLevel != "" {
		if level, err := logrus.ParseLevel(logLevel); err == nil {
			logger.SetLevel(level)
		}
	}

	logger.Info("启动SMTP中继服务器")

	// 连接MongoDB
	db, err := database.NewMongoDB(mongoURI, mongoDatabase, logger)
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
		URL:          rabbitmqURL,
		ExchangeName: rabbitmqExchange,
		QueueName:    rabbitmqQueue,
		RoutingKey:   "mail",
	}

	queueService, err := queue.NewService(queueConfig, db, logger)
	if err != nil {
		logger.WithError(err).Fatal("初始化队列服务失败")
	}
	defer queueService.Close()

	// 创建认证服务
	authService := auth.NewService(db, logger, secretKey)

	// 创建SMTP凭据服务
	credentialService := services.NewSMTPCredentialService(db, logger)

	// 创建SMTP服务器
	smtpConfig := &smtp.Config{
		Host:       smtpHost,
		Port25:     25,
		Port587:    587,
		Port465:    465,
		Domain:     smtpDomain,
		TLSCert:    getEnv("TLS_CERT_PATH", ""),
		TLSKey:     getEnv("TLS_KEY_PATH", ""),
		MaxMsgSize: 25 * 1024 * 1024, // 25MB
	}

	smtpServer := smtp.NewServer(smtpConfig, db, logger, authService, queueService, credentialService)

	// 启动SMTP服务器
	if err := smtpServer.Start(); err != nil {
		logger.WithError(err).Fatal("启动SMTP服务器失败")
	}
	defer smtpServer.Stop()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.Info("SMTP中继服务器启动完成")

	// 等待信号
	<-sigChan
	logger.Info("收到停止信号，正在关闭服务...")

	logger.Info("SMTP中继服务器已停止")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
