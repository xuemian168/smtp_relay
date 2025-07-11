package security

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Validator 安全验证器
type Validator struct {
	redis  *redis.Client
	logger *logrus.Logger
}

// NewValidator 创建安全验证器
func NewValidator(redis *redis.Client, logger *logrus.Logger) *Validator {
	return &Validator{
		redis:  redis,
		logger: logger,
	}
}

// ValidateEmailAddress 验证邮箱地址格式
func (v *Validator) ValidateEmailAddress(email string) error {
	if email == "" {
		return fmt.Errorf("邮箱地址不能为空")
	}

	// 基本格式检查
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("邮箱地址格式无效")
	}

	localPart := parts[0]
	domain := parts[1]

	// 检查本地部分
	if len(localPart) == 0 || len(localPart) > 64 {
		return fmt.Errorf("邮箱地址本地部分长度无效")
	}

	// 检查域名部分
	if len(domain) == 0 || len(domain) > 255 {
		return fmt.Errorf("邮箱地址域名部分长度无效")
	}

	// 检查域名格式
	if !strings.Contains(domain, ".") {
		return fmt.Errorf("邮箱地址域名格式无效")
	}

	return nil
}

// CheckRateLimit 检查发送频率限制
func (v *Validator) CheckRateLimit(userID primitive.ObjectID, remoteIP string) error {
	ctx := context.Background()

	// 检查用户发送频率
	userKey := fmt.Sprintf("rate_limit:user:%s", userID.Hex())
	userCount, err := v.redis.Incr(ctx, userKey).Result()
	if err != nil {
		v.logger.WithError(err).Error("检查用户发送频率失败")
		return err
	}

	if userCount == 1 {
		// 设置过期时间（1分钟）
		v.redis.Expire(ctx, userKey, time.Minute)
	}

	// 用户每分钟最多发送10封邮件
	if userCount > 10 {
		v.logger.WithFields(logrus.Fields{
			"user_id": userID.Hex(),
			"count":   userCount,
			"limit":   10,
		}).Warn("用户发送频率超过限制")
		return fmt.Errorf("发送频率过高，请稍后再试")
	}

	// 检查IP发送频率
	ipKey := fmt.Sprintf("rate_limit:ip:%s", remoteIP)
	ipCount, err := v.redis.Incr(ctx, ipKey).Result()
	if err != nil {
		v.logger.WithError(err).Error("检查IP发送频率失败")
		return err
	}

	if ipCount == 1 {
		// 设置过期时间（1分钟）
		v.redis.Expire(ctx, ipKey, time.Minute)
	}

	// 每个IP每分钟最多发送20封邮件
	if ipCount > 20 {
		v.logger.WithFields(logrus.Fields{
			"remote_ip": remoteIP,
			"count":     ipCount,
			"limit":     20,
		}).Warn("IP发送频率超过限制")
		return fmt.Errorf("IP发送频率过高，请稍后再试")
	}

	return nil
}

// CheckBlacklist 检查黑名单
func (v *Validator) CheckBlacklist(userID primitive.ObjectID, remoteIP string) error {
	ctx := context.Background()

	// 检查用户黑名单
	userBlacklistKey := fmt.Sprintf("blacklist:user:%s", userID.Hex())
	isBlacklisted, err := v.redis.Exists(ctx, userBlacklistKey).Result()
	if err != nil {
		v.logger.WithError(err).Error("检查用户黑名单失败")
		return err
	}

	if isBlacklisted > 0 {
		v.logger.WithField("user_id", userID.Hex()).Warn("用户在黑名单中")
		return fmt.Errorf("账户已被暂停使用")
	}

	// 检查IP黑名单
	ipBlacklistKey := fmt.Sprintf("blacklist:ip:%s", remoteIP)
	isIPBlacklisted, err := v.redis.Exists(ctx, ipBlacklistKey).Result()
	if err != nil {
		v.logger.WithError(err).Error("检查IP黑名单失败")
		return err
	}

	if isIPBlacklisted > 0 {
		v.logger.WithField("remote_ip", remoteIP).Warn("IP在黑名单中")
		return fmt.Errorf("IP地址已被限制访问")
	}

	return nil
}

// RecordFailedAttempt 记录失败尝试
func (v *Validator) RecordFailedAttempt(userID primitive.ObjectID, remoteIP string, reason string) {
	ctx := context.Background()

	// 记录用户失败尝试
	userFailKey := fmt.Sprintf("failed_attempts:user:%s", userID.Hex())
	userFailCount, err := v.redis.Incr(ctx, userFailKey).Result()
	if err != nil {
		v.logger.WithError(err).Error("记录用户失败尝试失败")
		return
	}

	if userFailCount == 1 {
		// 设置过期时间（1小时）
		v.redis.Expire(ctx, userFailKey, time.Hour)
	}

	// 用户1小时内失败超过5次，加入临时黑名单
	if userFailCount >= 5 {
		blacklistKey := fmt.Sprintf("blacklist:user:%s", userID.Hex())
		v.redis.Set(ctx, blacklistKey, reason, 24*time.Hour)
		v.logger.WithFields(logrus.Fields{
			"user_id":    userID.Hex(),
			"fail_count": userFailCount,
			"reason":     reason,
		}).Warn("用户因多次失败被加入黑名单")
	}

	// 记录IP失败尝试
	ipFailKey := fmt.Sprintf("failed_attempts:ip:%s", remoteIP)
	ipFailCount, err := v.redis.Incr(ctx, ipFailKey).Result()
	if err != nil {
		v.logger.WithError(err).Error("记录IP失败尝试失败")
		return
	}

	if ipFailCount == 1 {
		// 设置过期时间（1小时）
		v.redis.Expire(ctx, ipFailKey, time.Hour)
	}

	// IP 1小时内失败超过10次，加入临时黑名单
	if ipFailCount >= 10 {
		blacklistKey := fmt.Sprintf("blacklist:ip:%s", remoteIP)
		v.redis.Set(ctx, blacklistKey, reason, 24*time.Hour)
		v.logger.WithFields(logrus.Fields{
			"remote_ip":  remoteIP,
			"fail_count": ipFailCount,
			"reason":     reason,
		}).Warn("IP因多次失败被加入黑名单")
	}
}

// ValidateConnection 验证连接安全性
func (v *Validator) ValidateConnection(remoteAddr string) error {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return fmt.Errorf("无效的远程地址: %v", err)
	}

	// 检查是否为本地地址
	if v.isLocalAddress(host) {
		return nil // 允许本地连接
	}

	// 检查是否为已知的恶意IP段
	if v.isKnownBadIP(host) {
		v.logger.WithField("remote_ip", host).Warn("检测到已知恶意IP")
		return fmt.Errorf("连接被拒绝：IP地址不被信任")
	}

	return nil
}

// isLocalAddress 检查是否为本地地址
func (v *Validator) isLocalAddress(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	return ip.IsLoopback() || ip.IsPrivate()
}

// isKnownBadIP 检查是否为已知恶意IP
func (v *Validator) isKnownBadIP(host string) bool {
	// 这里可以集成第三方IP信誉数据库
	// 目前仅做基本检查
	badIPPrefixes := []string{
		"0.0.0.0",
		"255.255.255.255",
	}

	for _, prefix := range badIPPrefixes {
		if strings.HasPrefix(host, prefix) {
			return true
		}
	}

	return false
}

// LogSecurityEvent 记录安全事件
func (v *Validator) LogSecurityEvent(event string, userID primitive.ObjectID, remoteIP string, details map[string]interface{}) {
	logFields := logrus.Fields{
		"event":     event,
		"user_id":   userID.Hex(),
		"remote_ip": remoteIP,
		"timestamp": time.Now(),
	}

	for k, v := range details {
		logFields[k] = v
	}

	v.logger.WithFields(logFields).Warn("安全事件记录")
}
