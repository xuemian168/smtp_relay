// Package main SMTP中继服务API
// @title SMTP中继服务API
// @version 1.0
// @description 基于Go+Gin的SMTP转发中继服务API文档
// @contact.name API Support
// @contact.email support@smtp-relay.local
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token认证，格式: Bearer <token>
package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"smtp-relay/internal/api"
	"smtp-relay/internal/auth"
	"smtp-relay/internal/database"
	"smtp-relay/internal/services"
)

func main() {
	// 初始化日志
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	// 从环境变量读取配置
	mongoURI := getEnv("MONGODB_URI", "mongodb://localhost:27017/smtp_relay")
	mongoDatabase := getEnv("MONGODB_DATABASE", "smtp_relay")
	apiPort := getEnv("API_PORT", "8080")
	secretKey := getEnv("API_SECRET_KEY", "your-secret-key-change-in-production")

	// 设置日志级别
	if logLevel := getEnv("LOG_LEVEL", "info"); logLevel != "" {
		if level, err := logrus.ParseLevel(logLevel); err == nil {
			logger.SetLevel(level)
		}
	}

	logger.Info("启动SMTP中继API服务")

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

	// 创建认证服务
	authService := auth.NewService(db, logger, secretKey)

	// 创建SMTP凭据服务
	credentialService := services.NewSMTPCredentialService(db, logger)

	// 创建MailLog服务
	mailLogService := services.NewMailLogService(db, logger)

	// 创建API服务器
	apiConfig := &api.Config{
		Port:      apiPort,
		SecretKey: secretKey,
	}

	apiServer := api.NewServer(apiConfig, db, logger, authService, credentialService, mailLogService)

	// 启动API服务器
	go func() {
		if err := apiServer.Start(); err != nil {
			logger.WithError(err).Fatal("启动API服务器失败")
		}
	}()

	// 等待中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logger.WithField("port", apiPort).Info("SMTP中继API服务启动完成")

	// 等待信号
	<-sigChan
	logger.Info("收到停止信号，正在关闭服务...")

	// 停止API服务器
	if err := apiServer.Stop(); err != nil {
		logger.WithError(err).Error("停止API服务器失败")
	}

	logger.Info("SMTP中继API服务已停止")
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
