package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"smtp-relay/internal/database"
	"smtp-relay/internal/models"
)

// Service 认证服务
type Service struct {
	db     *database.MongoDB
	logger *logrus.Logger
	secret []byte
}

// NewService 创建认证服务
func NewService(db *database.MongoDB, logger *logrus.Logger, secret string) *Service {
	return &Service{
		db:     db,
		logger: logger,
		secret: []byte(secret),
	}
}

// AuthenticateSMTP 验证SMTP用户凭据（使用新的多密钥对系统）
func (s *Service) AuthenticateSMTP(username, password string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查找SMTP凭据
	var credential models.SMTPCredential
	credentialCollection := s.db.GetCollection("smtp_credentials")

	filter := bson.M{
		"username": username,
		"active":   true,
	}

	err := credentialCollection.FindOne(ctx, filter).Decode(&credential)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("SMTP凭据不存在")
		}
		s.logger.WithError(err).Error("查询SMTP凭据失败")
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(credential.PasswordHash), []byte(password)); err != nil {
		s.logger.WithField("username", username).Warn("SMTP密码验证失败")
		return nil, errors.New("密码错误")
	}

	// 查找对应的用户
	var user models.User
	userCollection := s.db.GetCollection("users")

	userFilter := bson.M{
		"_id":    credential.UserID,
		"status": "active",
	}

	err = userCollection.FindOne(ctx, userFilter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("用户不存在")
		}
		s.logger.WithError(err).Error("查询用户失败")
		return nil, err
	}

	// 更新凭据的最后使用时间
	updateFilter := bson.M{"_id": credential.ID}
	update := bson.M{
		"$set": bson.M{
			"last_used": time.Now(),
		},
		"$inc": bson.M{
			"usage_count": 1,
		},
	}

	_, err = credentialCollection.UpdateOne(ctx, updateFilter, update)
	if err != nil {
		s.logger.WithError(err).Warn("更新SMTP凭据使用记录失败")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":       user.ID.Hex(),
		"credential_id": credential.ID.Hex(),
		"username":      username,
	}).Info("SMTP认证成功")

	return &user, nil
}

// CheckQuota 检查用户配额（已废弃，使用CheckCredentialQuota）
func (s *Service) CheckQuota(userID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取用户配额信息
	quotaCollection := s.db.GetCollection("user_quotas")
	today := time.Now().Truncate(24 * time.Hour)

	var quota models.UserQuota
	filter := bson.M{
		"user_id": userID,
		"date":    today,
	}

	err := quotaCollection.FindOne(ctx, filter).Decode(&quota)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新的配额记录
			quota = models.UserQuota{
				UserID:        userID,
				Date:          today,
				DailyCount:    0,
				DailyLimit:    1000, // 默认值
				HourlyCount:   0,
				HourlyLimit:   100, // 默认值
				LastResetHour: time.Now().Truncate(time.Hour),
				LastResetDay:  today,
			}

			if _, err := quotaCollection.InsertOne(ctx, quota); err != nil {
				s.logger.WithError(err).Error("创建配额记录失败")
				return err
			}
		} else {
			s.logger.WithError(err).Error("查询配额失败")
			return err
		}
	}

	// 检查小时配额
	currentHour := time.Now().Truncate(time.Hour)
	if quota.LastResetHour.Before(currentHour) {
		// 重置小时计数
		quota.HourlyCount = 0
		quota.LastResetHour = currentHour
	}

	if quota.HourlyCount >= quota.HourlyLimit {
		return errors.New("小时发送配额已用完")
	}

	// 检查日配额
	if quota.DailyCount >= quota.DailyLimit {
		return errors.New("日发送配额已用完")
	}

	// 更新配额计数
	update := bson.M{
		"$inc": bson.M{
			"daily_count":  1,
			"hourly_count": 1,
		},
		"$set": bson.M{
			"last_reset_hour": quota.LastResetHour,
		},
	}

	_, err = quotaCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.WithError(err).Error("更新配额失败")
		return err
	}

	return nil
}

// CheckCredentialQuota 检查SMTP凭据配额
func (s *Service) CheckCredentialQuota(credentialID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 获取凭据信息
	var credential models.SMTPCredential
	credentialCollection := s.db.GetCollection("smtp_credentials")

	err := credentialCollection.FindOne(ctx, bson.M{"_id": credentialID}).Decode(&credential)
	if err != nil {
		s.logger.WithError(err).Error("查询SMTP凭据失败")
		return err
	}

	// 获取凭据配额信息
	quotaCollection := s.db.GetCollection("credential_quotas")
	today := time.Now().Truncate(24 * time.Hour)

	var quota models.CredentialQuota
	filter := bson.M{
		"credential_id": credentialID,
		"date":          today,
	}

	err = quotaCollection.FindOne(ctx, filter).Decode(&quota)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// 创建新的配额记录
			quota = models.CredentialQuota{
				CredentialID:  credentialID,
				Date:          today,
				DailyCount:    0,
				DailyLimit:    credential.Settings.DailyQuota,
				HourlyCount:   0,
				HourlyLimit:   credential.Settings.HourlyQuota,
				LastResetHour: time.Now().Truncate(time.Hour),
				LastResetDay:  today,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			if _, err := quotaCollection.InsertOne(ctx, quota); err != nil {
				s.logger.WithError(err).Error("创建凭据配额记录失败")
				return err
			}
		} else {
			s.logger.WithError(err).Error("查询凭据配额失败")
			return err
		}
	}

	// 检查小时配额
	currentHour := time.Now().Truncate(time.Hour)
	if quota.LastResetHour.Before(currentHour) {
		// 重置小时计数
		quota.HourlyCount = 0
		quota.LastResetHour = currentHour
	}

	if quota.HourlyCount >= quota.HourlyLimit {
		return errors.New("小时发送配额已用完")
	}

	// 检查日配额
	if quota.DailyCount >= quota.DailyLimit {
		return errors.New("日发送配额已用完")
	}

	// 更新配额计数
	update := bson.M{
		"$inc": bson.M{
			"daily_count":  1,
			"hourly_count": 1,
		},
		"$set": bson.M{
			"last_reset_hour": quota.LastResetHour,
		},
	}

	_, err = quotaCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.WithError(err).Error("更新凭据配额失败")
		return err
	}

	return nil
}

// GenerateJWT 生成JWT令牌
func (s *Service) GenerateJWT(userID primitive.ObjectID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.Hex(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// ValidateJWT 验证JWT令牌
func (s *Service) ValidateJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方法")
		}
		return s.secret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userIDStr := claims["user_id"].(string)
		return userIDStr, nil
	}

	return "", errors.New("无效的令牌")
}

// AuthenticateUser 验证用户邮箱和密码
func (s *Service) AuthenticateUser(email, password string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查找用户
	var user models.User
	collection := s.db.GetCollection("users")

	filter := bson.M{
		"email":  email,
		"status": "active",
	}

	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("用户不存在")
		}
		s.logger.WithError(err).Error("查询用户失败")
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.WithField("email", email).Warn("密码验证失败")
		return nil, errors.New("密码错误")
	}

	s.logger.WithField("user_id", user.ID.Hex()).Info("用户认证成功")
	return &user, nil
}

// AuthenticateUserByUsername 验证用户名和密码
func (s *Service) AuthenticateUserByUsername(username, password string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 查找用户
	var user models.User
	collection := s.db.GetCollection("users")

	filter := bson.M{
		"username": username,
		"status":   "active",
	}

	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("用户不存在")
		}
		s.logger.WithError(err).Error("查询用户失败")
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		s.logger.WithField("username", username).Warn("密码验证失败")
		return nil, errors.New("密码错误")
	}

	s.logger.WithField("user_id", user.ID.Hex()).Info("用户认证成功（用户名登录）")
	return &user, nil
}

// CreateUser 创建新用户
func (s *Service) CreateUser(username, email, password string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 检查用户名是否已存在
	collection := s.db.GetCollection("users")
	count, err := collection.CountDocuments(ctx, bson.M{"username": username})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	count, err = collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, errors.New("邮箱已存在")
	}

	// 加密密码
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(passwordHash),
		Status:       "active",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Settings: models.UserSettings{
			DailyQuota:     1000,
			HourlyQuota:    100,
			AllowedDomains: []string{},
		},
	}

	result, err := collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	s.logger.WithField("user_id", user.ID.Hex()).Info("用户创建成功")

	return user, nil
}

// GetUserByID 根据ID获取用户
func (s *Service) GetUserByID(userIDStr string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := primitive.ObjectIDFromHex(userIDStr)
	if err != nil {
		return nil, errors.New("无效的用户ID")
	}

	var user models.User
	collection := s.db.GetCollection("users")

	err = collection.FindOne(ctx, bson.M{"_id": userID, "status": "active"}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	return &user, nil
}

// HashPassword 密码哈希
func (s *Service) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
