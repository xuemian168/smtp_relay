package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DKIMKeyPair DKIM密钥对模型
type DKIMKeyPair struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Domain       string             `bson:"domain" json:"domain"`                         // 签名域名
	Selector     string             `bson:"selector" json:"selector"`                     // DKIM选择器
	PrivateKey   string             `bson:"private_key" json:"-"`                         // 私钥（不返回给前端）
	PublicKey    string             `bson:"public_key" json:"public_key"`                 // 公钥
	KeySize      int                `bson:"key_size" json:"key_size"`                     // 密钥长度（1024, 2048）
	Algorithm    string             `bson:"algorithm" json:"algorithm"`                   // 签名算法（rsa-sha256）
	Status       string             `bson:"status" json:"status"`                         // active, inactive, expired
	DNSRecord    string             `bson:"dns_record" json:"dns_record"`                 // 生成的DNS TXT记录
	DNSVerified  bool               `bson:"dns_verified" json:"dns_verified"`             // DNS记录是否已验证
	LastVerified *time.Time         `bson:"last_verified,omitempty" json:"last_verified"` // 最后验证时间
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	ExpiresAt    *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"` // 密钥过期时间
}

// DKIMSettings DKIM配置设置
type DKIMSettings struct {
	Enabled        bool                `bson:"enabled" json:"enabled"`                                   // 是否启用DKIM
	DefaultKeyID   *primitive.ObjectID `bson:"default_key_id,omitempty" json:"default_key_id,omitempty"` // 默认密钥ID
	SigningDomains []string            `bson:"signing_domains" json:"signing_domains"`                   // 允许签名的域名列表
	HeaderCanon    string              `bson:"header_canon" json:"header_canon"`                         // 头部规范化（relaxed/simple）
	BodyCanon      string              `bson:"body_canon" json:"body_canon"`                             // 正文规范化（relaxed/simple）
	SignHeaders    []string            `bson:"sign_headers" json:"sign_headers"`                         // 要签名的头部列表
}

// DKIMConfig DKIM域名配置
type DKIMConfig struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id"`
	Domain    string             `bson:"domain" json:"domain"`           // 配置的域名
	KeyPairID primitive.ObjectID `bson:"key_pair_id" json:"key_pair_id"` // 关联的密钥对ID
	Settings  DKIMSettings       `bson:"settings" json:"settings"`       // DKIM设置
	Active    bool               `bson:"active" json:"active"`           // 是否激活
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// DNSRecord DNS记录结构
type DNSRecord struct {
	Type     string `json:"type"`     // TXT
	Name     string `json:"name"`     // 记录名称，如 selector._domainkey.example.com
	Value    string `json:"value"`    // 记录值
	TTL      int    `json:"ttl"`      // TTL值
	Priority int    `json:"priority"` // 优先级（对TXT记录通常为0）
}

// DKIMValidationResult DKIM验证结果
type DKIMValidationResult struct {
	Domain       string    `json:"domain"`
	Selector     string    `json:"selector"`
	Valid        bool      `json:"valid"`
	DNSFound     bool      `json:"dns_found"`
	DNSRecord    string    `json:"dns_record,omitempty"`
	ExpectedDNS  string    `json:"expected_dns"`
	ErrorMessage string    `json:"error_message,omitempty"`
	CheckedAt    time.Time `json:"checked_at"`
}

// GetDNSRecordName 获取DNS记录名称
func (d *DKIMKeyPair) GetDNSRecordName() string {
	return d.Selector + "._domainkey." + d.Domain
}

// GetDNSRecordValue 获取DNS记录值
func (d *DKIMKeyPair) GetDNSRecordValue() string {
	return d.DNSRecord
}

// IsExpired 检查密钥是否过期
func (d *DKIMKeyPair) IsExpired() bool {
	if d.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*d.ExpiresAt)
}

// NeedsVerification 检查是否需要重新验证DNS
func (d *DKIMKeyPair) NeedsVerification() bool {
	if !d.DNSVerified {
		return true
	}
	if d.LastVerified == nil {
		return true
	}
	// 每24小时重新验证一次
	return time.Since(*d.LastVerified) > 24*time.Hour
}
