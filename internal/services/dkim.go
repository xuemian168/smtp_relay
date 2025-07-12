package services

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"net"
	"strings"
	"time"

	"smtp-relay/internal/database"
	"smtp-relay/internal/models"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// DKIMService DKIM服务
type DKIMService struct {
	db     *database.MongoDB
	logger *logrus.Logger
}

// NewDKIMService 创建DKIM服务实例
func NewDKIMService(db *database.MongoDB, logger *logrus.Logger) *DKIMService {
	return &DKIMService{
		db:     db,
		logger: logger,
	}
}

// GenerateKeyPair 生成DKIM密钥对
func (s *DKIMService) GenerateKeyPair(userID primitive.ObjectID, domain, selector string, keySize int) (*models.DKIMKeyPair, error) {
	// 验证参数
	if domain == "" || selector == "" {
		return nil, fmt.Errorf("域名和选择器不能为空")
	}
	if keySize != 1024 && keySize != 2048 {
		keySize = 2048 // 默认使用2048位
	}

	// 检查是否已存在相同的域名+选择器组合
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("dkim_keys")
	existingCount, err := collection.CountDocuments(ctx, bson.M{
		"user_id":  userID,
		"domain":   domain,
		"selector": selector,
		"status":   bson.M{"$ne": "deleted"},
	})
	if err != nil {
		return nil, fmt.Errorf("检查现有密钥失败: %w", err)
	}
	if existingCount > 0 {
		return nil, fmt.Errorf("该域名和选择器的密钥对已存在")
	}

	// 生成RSA密钥对
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("生成RSA密钥失败: %w", err)
	}

	// 编码私钥
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	}
	privateKeyStr := string(pem.EncodeToMemory(privateKeyPEM))

	// 编码公钥
	publicKeyDER, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("编码公钥失败: %w", err)
	}
	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyDER,
	}
	publicKeyStr := string(pem.EncodeToMemory(publicKeyPEM))

	// 生成DNS记录
	dnsRecord := s.generateDNSRecord(&privateKey.PublicKey)

	// 创建DKIM密钥对记录
	keyPair := &models.DKIMKeyPair{
		ID:          primitive.NewObjectID(),
		UserID:      userID,
		Domain:      domain,
		Selector:    selector,
		PrivateKey:  privateKeyStr,
		PublicKey:   publicKeyStr,
		KeySize:     keySize,
		Algorithm:   "rsa-sha256",
		Status:      "active",
		DNSRecord:   dnsRecord,
		DNSVerified: false,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 保存到数据库
	_, err = collection.InsertOne(ctx, keyPair)
	if err != nil {
		return nil, fmt.Errorf("保存DKIM密钥对失败: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":  userID.Hex(),
		"domain":   domain,
		"selector": selector,
		"key_size": keySize,
	}).Info("DKIM密钥对生成成功")

	return keyPair, nil
}

// generateDNSRecord 生成DNS TXT记录值
func (s *DKIMService) generateDNSRecord(publicKey *rsa.PublicKey) string {
	// 提取公钥的模数和指数
	publicKeyDER, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		s.logger.WithError(err).Error("编码公钥失败")
		return ""
	}

	// Base64编码公钥
	publicKeyB64 := base64.StdEncoding.EncodeToString(publicKeyDER)

	// 构建DNS记录
	// 格式: v=DKIM1; h=sha256; k=rsa; p=<base64-encoded-public-key>
	return fmt.Sprintf("v=DKIM1; h=sha256; k=rsa; p=%s", publicKeyB64)
}

// ListKeyPairs 获取用户的DKIM密钥对列表
func (s *DKIMService) ListKeyPairs(userID primitive.ObjectID) ([]*models.DKIMKeyPair, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("dkim_keys")
	filter := bson.M{
		"user_id": userID,
		"status":  bson.M{"$ne": "deleted"},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询DKIM密钥对失败: %w", err)
	}
	defer cursor.Close(ctx)

	var keyPairs []*models.DKIMKeyPair
	if err := cursor.All(ctx, &keyPairs); err != nil {
		return nil, fmt.Errorf("解析DKIM密钥对失败: %w", err)
	}

	return keyPairs, nil
}

// GetKeyPair 获取指定的DKIM密钥对
func (s *DKIMService) GetKeyPair(userID, keyPairID primitive.ObjectID) (*models.DKIMKeyPair, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("dkim_keys")
	filter := bson.M{
		"_id":     keyPairID,
		"user_id": userID,
		"status":  bson.M{"$ne": "deleted"},
	}

	var keyPair models.DKIMKeyPair
	err := collection.FindOne(ctx, filter).Decode(&keyPair)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("DKIM密钥对不存在")
		}
		return nil, fmt.Errorf("获取DKIM密钥对失败: %w", err)
	}

	return &keyPair, nil
}

// DeleteKeyPair 删除DKIM密钥对
func (s *DKIMService) DeleteKeyPair(userID, keyPairID primitive.ObjectID) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("dkim_keys")
	filter := bson.M{
		"_id":     keyPairID,
		"user_id": userID,
	}

	update := bson.M{
		"$set": bson.M{
			"status":     "deleted",
			"updated_at": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("删除DKIM密钥对失败: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("DKIM密钥对不存在")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":     userID.Hex(),
		"key_pair_id": keyPairID.Hex(),
	}).Info("DKIM密钥对删除成功")

	return nil
}

// VerifyDNSRecord 验证DNS记录
func (s *DKIMService) VerifyDNSRecord(userID, keyPairID primitive.ObjectID) (*models.DKIMValidationResult, error) {
	// 获取密钥对信息
	keyPair, err := s.GetKeyPair(userID, keyPairID)
	if err != nil {
		return nil, err
	}

	result := &models.DKIMValidationResult{
		Domain:      keyPair.Domain,
		Selector:    keyPair.Selector,
		ExpectedDNS: keyPair.DNSRecord,
		CheckedAt:   time.Now(),
	}

	// 查询DNS记录
	dnsName := keyPair.GetDNSRecordName()
	txtRecords, err := net.LookupTXT(dnsName)
	if err != nil {
		result.Valid = false
		result.DNSFound = false
		result.ErrorMessage = fmt.Sprintf("DNS查询失败: %v", err)
		return result, nil
	}

	// 检查是否找到匹配的记录
	result.DNSFound = len(txtRecords) > 0
	for _, record := range txtRecords {
		// 移除空格和换行符进行比较
		cleanRecord := strings.ReplaceAll(strings.ReplaceAll(record, " ", ""), "\n", "")
		cleanExpected := strings.ReplaceAll(strings.ReplaceAll(keyPair.DNSRecord, " ", ""), "\n", "")

		if cleanRecord == cleanExpected {
			result.Valid = true
			result.DNSRecord = record
			break
		}
	}

	if result.DNSFound && !result.Valid {
		result.ErrorMessage = "DNS记录存在但内容不匹配"
		if len(txtRecords) > 0 {
			result.DNSRecord = txtRecords[0]
		}
	} else if !result.DNSFound {
		result.ErrorMessage = "未找到DNS记录"
	}

	// 更新数据库中的验证状态
	if result.Valid {
		s.updateVerificationStatus(keyPairID, true, time.Now())
	}

	return result, nil
}

// updateVerificationStatus 更新验证状态
func (s *DKIMService) updateVerificationStatus(keyPairID primitive.ObjectID, verified bool, verifiedAt time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("dkim_keys")
	filter := bson.M{"_id": keyPairID}

	update := bson.M{
		"$set": bson.M{
			"dns_verified":  verified,
			"last_verified": verifiedAt,
			"updated_at":    time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

// GetDNSRecord 获取DNS记录信息
func (s *DKIMService) GetDNSRecord(userID, keyPairID primitive.ObjectID) (*models.DNSRecord, error) {
	keyPair, err := s.GetKeyPair(userID, keyPairID)
	if err != nil {
		return nil, err
	}

	return &models.DNSRecord{
		Type:     "TXT",
		Name:     keyPair.GetDNSRecordName(),
		Value:    keyPair.GetDNSRecordValue(),
		TTL:      3600,
		Priority: 0,
	}, nil
}

// GetKeyPairsByDomain 根据域名获取DKIM密钥对
func (s *DKIMService) GetKeyPairsByDomain(userID primitive.ObjectID, domain string) ([]*models.DKIMKeyPair, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := s.db.GetCollection("dkim_keys")
	filter := bson.M{
		"user_id": userID,
		"domain":  domain,
		"status":  "active",
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("查询域名DKIM密钥对失败: %w", err)
	}
	defer cursor.Close(ctx)

	var keyPairs []*models.DKIMKeyPair
	if err := cursor.All(ctx, &keyPairs); err != nil {
		return nil, fmt.Errorf("解析域名DKIM密钥对失败: %w", err)
	}

	return keyPairs, nil
}

// RotateKeyPair 轮换密钥对
func (s *DKIMService) RotateKeyPair(userID, keyPairID primitive.ObjectID) (*models.DKIMKeyPair, error) {
	// 获取现有密钥对
	oldKeyPair, err := s.GetKeyPair(userID, keyPairID)
	if err != nil {
		return nil, err
	}

	// 生成新的选择器（在原选择器后加上时间戳）
	newSelector := fmt.Sprintf("%s-%d", oldKeyPair.Selector, time.Now().Unix())

	// 生成新密钥对
	newKeyPair, err := s.GenerateKeyPair(userID, oldKeyPair.Domain, newSelector, oldKeyPair.KeySize)
	if err != nil {
		return nil, fmt.Errorf("生成新密钥对失败: %w", err)
	}

	// 标记旧密钥对为即将过期
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := s.db.GetCollection("dkim_keys")
	expiresAt := time.Now().Add(30 * 24 * time.Hour) // 30天后过期

	_, err = collection.UpdateOne(ctx,
		bson.M{"_id": keyPairID},
		bson.M{
			"$set": bson.M{
				"status":     "expiring",
				"expires_at": expiresAt,
				"updated_at": time.Now(),
			},
		},
	)
	if err != nil {
		s.logger.WithError(err).Error("更新旧密钥对状态失败")
	}

	s.logger.WithFields(logrus.Fields{
		"user_id":      userID.Hex(),
		"old_key_id":   keyPairID.Hex(),
		"new_key_id":   newKeyPair.ID.Hex(),
		"domain":       oldKeyPair.Domain,
		"old_selector": oldKeyPair.Selector,
		"new_selector": newSelector,
	}).Info("DKIM密钥对轮换成功")

	return newKeyPair, nil
}
