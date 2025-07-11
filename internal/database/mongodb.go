package database

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB 数据库连接结构
type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
	logger   *logrus.Logger
}

// NewMongoDB 创建MongoDB连接
func NewMongoDB(uri, dbName string, logger *logrus.Logger) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 设置连接选项
	clientOptions := options.Client().ApplyURI(uri)
	clientOptions.SetMaxPoolSize(100)
	clientOptions.SetMinPoolSize(10)
	clientOptions.SetMaxConnIdleTime(30 * time.Second)

	// 连接MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.WithError(err).Error("连接MongoDB失败")
		return nil, err
	}

	// 验证连接
	if err := client.Ping(ctx, nil); err != nil {
		logger.WithError(err).Error("MongoDB连接验证失败")
		return nil, err
	}

	database := client.Database(dbName)
	logger.Info("MongoDB连接成功")

	return &MongoDB{
		Client:   client,
		Database: database,
		logger:   logger,
	}, nil
}

// Close 关闭数据库连接
func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := m.Client.Disconnect(ctx); err != nil {
		m.logger.WithError(err).Error("关闭MongoDB连接失败")
		return err
	}

	m.logger.Info("MongoDB连接已关闭")
	return nil
}

// GetCollection 获取集合
func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}

// CreateIndexes 创建索引
func (m *MongoDB) CreateIndexes() error {
	ctx := context.Background()

	// 用户集合索引
	userCollection := m.GetCollection("users")
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{"email", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"status", 1}},
		},
	}

	if _, err := userCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return err
	}

	// SMTP凭据集合索引
	credentialCollection := m.GetCollection("smtp_credentials")
	credentialIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"username", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"user_id", 1}},
		},
		{
			Keys: bson.D{{"active", 1}},
		},
	}

	if _, err := credentialCollection.Indexes().CreateMany(ctx, credentialIndexes); err != nil {
		return err
	}

	// MailLog集合索引
	mailLogCollection := m.GetCollection("mail_logs")
	mailLogIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{"user_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"credential_id", 1}, {"created_at", -1}},
		},
		{
			Keys: bson.D{{"message_id", 1}},
		},
		{
			Keys: bson.D{{"status", 1}},
		},
		{
			Keys: bson.D{{"created_at", -1}},
		},
	}

	if _, err := mailLogCollection.Indexes().CreateMany(ctx, mailLogIndexes); err != nil {
		return err
	}

	// 用户配额集合索引
	quotaCollection := m.GetCollection("user_quotas")
	quotaIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"user_id", 1}, {"date", 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := quotaCollection.Indexes().CreateMany(ctx, quotaIndexes); err != nil {
		return err
	}

	// 凭据配额集合索引
	credentialQuotaCollection := m.GetCollection("credential_quotas")
	credentialQuotaIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"credential_id", 1}, {"date", 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := credentialQuotaCollection.Indexes().CreateMany(ctx, credentialQuotaIndexes); err != nil {
		return err
	}

	// IP信誉集合索引
	ipCollection := m.GetCollection("ip_reputation")
	ipIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"ip", 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := ipCollection.Indexes().CreateMany(ctx, ipIndexes); err != nil {
		return err
	}

	// SMTP配置集合索引
	smtpConfigCollection := m.GetCollection("smtp_configs")
	smtpConfigIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{"host", 1}, {"port", 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{"active", 1}},
		},
		{
			Keys: bson.D{{"priority", 1}},
		},
	}

	if _, err := smtpConfigCollection.Indexes().CreateMany(ctx, smtpConfigIndexes); err != nil {
		return err
	}

	m.logger.Info("MongoDB索引创建完成")
	return nil
}
