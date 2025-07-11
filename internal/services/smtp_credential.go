package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"smtp-relay/internal/database"
	"smtp-relay/internal/models"
)

// SMTPCredentialService SMTP凭据管理服务
type SMTPCredentialService struct {
	db     *database.MongoDB
	logger *logrus.Logger
}

// NewSMTPCredentialService 创建SMTP凭据管理服务
func NewSMTPCredentialService(db *database.MongoDB, logger *logrus.Logger) *SMTPCredentialService {
	return &SMTPCredentialService{
		db:     db,
		logger: logger,
	}
}

// CreateCredential 创建新的SMTP凭据
func (s *SMTPCredentialService) CreateCredential(userID primitive.ObjectID, name, description string) (*models.SMTPCredential, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查用户是否存在
	userCollection := s.db.GetCollection("users")
	var user models.User
	err := userCollection.FindOne(ctx, bson.M{"_id": userID, "status": "active"}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, "", errors.New("用户不存在或已被禁用")
		}
		return nil, "", err
	}

	// 检查凭据数量限制
	credentialCollection := s.db.GetCollection("smtp_credentials")
	count, err := credentialCollection.CountDocuments(ctx, bson.M{"user_id": userID, "status": "active"})
	if err != nil {
		return nil, "", err
	}

	if count >= 10 { // 每个用户最多10个凭据
		return nil, "", errors.New("SMTP凭据数量已达到上限（10个）")
	}

	// 生成唯一的SMTP用户名
	smtpUsername := s.generateSMTPUsername(userID)

	// 生成随机密码
	password := s.generateRandomPassword()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	// 创建SMTP凭据
	credential := &models.SMTPCredential{
		UserID:       userID,
		Name:         name,
		Username:     smtpUsername,
		PasswordHash: string(passwordHash),
		Description:  description,
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		UsageCount:   0,
		Settings: models.SMTPCredentialSettings{
			DailyQuota:     user.Settings.DailyQuota,     // 继承用户设置
			HourlyQuota:    user.Settings.HourlyQuota,    // 继承用户设置
			AllowedDomains: user.Settings.AllowedDomains, // 继承用户设置
			MaxRecipients:  100,
		},
	}

	result, err := credentialCollection.InsertOne(ctx, credential)
	if err != nil {
		return nil, "", err
	}

	credential.ID = result.InsertedID.(primitive.ObjectID)

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID.Hex(),
		"credential_id": credential.ID.Hex(),
		"name":          name,
		"username":      smtpUsername,
	}).Info("创建SMTP凭据成功")

	return credential, password, nil
}

// ListCredentials 获取用户的SMTP凭据列表
func (s *SMTPCredentialService) ListCredentials(userID primitive.ObjectID) ([]*models.SMTPCredential, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("smtp_credentials")
	filter := bson.M{"user_id": userID, "status": "active"}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var credentials []*models.SMTPCredential
	for cursor.Next(ctx) {
		var credential models.SMTPCredential
		if err := cursor.Decode(&credential); err != nil {
			continue
		}
		credentials = append(credentials, &credential)
	}

	return credentials, nil
}

// GetCredential 获取单个SMTP凭据
func (s *SMTPCredentialService) GetCredential(userID, credentialID primitive.ObjectID) (*models.SMTPCredential, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("smtp_credentials")
	filter := bson.M{
		"_id":     credentialID,
		"user_id": userID,
		"status":  "active",
	}

	var credential models.SMTPCredential
	err := collection.FindOne(ctx, filter).Decode(&credential)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("SMTP凭据不存在")
		}
		return nil, err
	}

	return &credential, nil
}

// UpdateCredential 更新SMTP凭据
func (s *SMTPCredentialService) UpdateCredential(userID, credentialID primitive.ObjectID, name, description string, settings models.SMTPCredentialSettings) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("smtp_credentials")
	filter := bson.M{
		"_id":     credentialID,
		"user_id": userID,
		"status":  "active",
	}

	update := bson.M{
		"$set": bson.M{
			"name":        name,
			"description": description,
			"settings":    settings,
			"updated_at":  time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("SMTP凭据不存在")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID.Hex(),
		"credential_id": credentialID.Hex(),
		"name":          name,
	}).Info("更新SMTP凭据成功")

	return nil
}

// ResetPassword 重置SMTP凭据密码
func (s *SMTPCredentialService) ResetPassword(userID, credentialID primitive.ObjectID) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("smtp_credentials")
	filter := bson.M{
		"_id":     credentialID,
		"user_id": userID,
		"status":  "active",
	}

	// 生成新密码
	newPassword := s.generateRandomPassword()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	update := bson.M{
		"$set": bson.M{
			"password_hash": string(passwordHash),
			"updated_at":    time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return "", err
	}

	if result.MatchedCount == 0 {
		return "", errors.New("SMTP凭据不存在")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID.Hex(),
		"credential_id": credentialID.Hex(),
	}).Info("重置SMTP凭据密码成功")

	return newPassword, nil
}

// DeleteCredential 删除SMTP凭据
func (s *SMTPCredentialService) DeleteCredential(userID, credentialID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("smtp_credentials")
	filter := bson.M{
		"_id":     credentialID,
		"user_id": userID,
		"status":  "active",
	}

	update := bson.M{
		"$set": bson.M{
			"status":     "deleted",
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("SMTP凭据不存在")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       userID.Hex(),
		"credential_id": credentialID.Hex(),
	}).Info("删除SMTP凭据成功")

	return nil
}

// AuthenticateSMTP 验证SMTP凭据
func (s *SMTPCredentialService) AuthenticateSMTP(username, password string) (*models.SMTPCredential, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("smtp_credentials")
	filter := bson.M{
		"username": username,
		"status":   "active",
	}

	var credential models.SMTPCredential
	err := collection.FindOne(ctx, filter).Decode(&credential)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("SMTP凭据不存在")
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(credential.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("密码错误")
	}

	// 更新使用统计
	go s.updateUsageStats(credential.ID)

	return &credential, nil
}

// updateUsageStats 更新使用统计
func (s *SMTPCredentialService) updateUsageStats(credentialID primitive.ObjectID) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("smtp_credentials")
	filter := bson.M{"_id": credentialID}
	update := bson.M{
		"$inc": bson.M{"usage_count": 1},
		"$set": bson.M{"last_used": time.Now()},
	}

	collection.UpdateOne(ctx, filter, update)
}

// generateSMTPUsername 生成SMTP用户名
func (s *SMTPCredentialService) generateSMTPUsername(userID primitive.ObjectID) string {
	// 生成格式：relay_{userID前8位}_{随机4位}
	userIDStr := userID.Hex()
	randomBytes := make([]byte, 2)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)

	return fmt.Sprintf("relay_%s_%s", userIDStr[:8], randomStr)
}

// generateRandomPassword 生成随机密码
func (s *SMTPCredentialService) generateRandomPassword() string {
	// 生成16位随机密码
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
