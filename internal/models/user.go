package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User 用户信息结构
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Status       string             `bson:"status" json:"status"` // active, suspended, deleted
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `bson:"updated_at" json:"updated_at"`
	Settings     UserSettings       `bson:"settings" json:"settings"`
}

// SMTPCredential SMTP认证凭据（支持多个密钥对）
type SMTPCredential struct {
	ID           primitive.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID     `bson:"user_id" json:"user_id"`
	Name         string                 `bson:"name" json:"name"`               // 凭据名称，如"mailcow-server1"
	Username     string                 `bson:"username" json:"username"`       // SMTP用户名
	PasswordHash string                 `bson:"password_hash" json:"-"`         // SMTP密码哈希
	Description  string                 `bson:"description" json:"description"` // 描述信息
	Status       string                 `bson:"status" json:"status"`           // active, disabled
	CreatedAt    time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time              `bson:"updated_at" json:"updated_at"`
	LastUsed     *time.Time             `bson:"last_used,omitempty" json:"last_used,omitempty"`
	UsageCount   int64                  `bson:"usage_count" json:"usage_count"` // 使用次数
	Settings     SMTPCredentialSettings `bson:"settings" json:"settings"`
}

// SMTPCredentialSettings SMTP凭据设置
type SMTPCredentialSettings struct {
	DailyQuota     int      `bson:"daily_quota" json:"daily_quota"`         // 该凭据的日配额
	HourlyQuota    int      `bson:"hourly_quota" json:"hourly_quota"`       // 该凭据的小时配额
	AllowedDomains []string `bson:"allowed_domains" json:"allowed_domains"` // 允许发送的域名
	MaxRecipients  int      `bson:"max_recipients" json:"max_recipients"`   // 单封邮件最大收件人数
}

// UserSettings 用户设置
type UserSettings struct {
	DailyQuota     int      `bson:"daily_quota" json:"daily_quota"`
	HourlyQuota    int      `bson:"hourly_quota" json:"hourly_quota"`
	AllowedDomains []string `bson:"allowed_domains" json:"allowed_domains"`
}
