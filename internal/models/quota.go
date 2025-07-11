package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserQuota 用户配额统计
type UserQuota struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID        primitive.ObjectID `bson:"user_id" json:"user_id"`
	Date          time.Time          `bson:"date" json:"date"`
	DailyCount    int                `bson:"daily_count" json:"daily_count"`
	DailyLimit    int                `bson:"daily_limit" json:"daily_limit"`
	HourlyCount   int                `bson:"hourly_count" json:"hourly_count"`
	HourlyLimit   int                `bson:"hourly_limit" json:"hourly_limit"`
	LastResetHour time.Time          `bson:"last_reset_hour" json:"last_reset_hour"`
	LastResetDay  time.Time          `bson:"last_reset_day" json:"last_reset_day"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Key   string             `bson:"key" json:"key"`
	Value interface{}        `bson:"value" json:"value"`
	Type  string             `bson:"type" json:"type"` // string, int, bool, float
}

// CredentialQuota SMTP凭据配额统计
type CredentialQuota struct {
	ID            primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	CredentialID  primitive.ObjectID `bson:"credential_id" json:"credential_id"`
	Date          time.Time          `bson:"date" json:"date"`
	DailyCount    int                `bson:"daily_count" json:"daily_count"`
	DailyLimit    int                `bson:"daily_limit" json:"daily_limit"`
	HourlyCount   int                `bson:"hourly_count" json:"hourly_count"`
	HourlyLimit   int                `bson:"hourly_limit" json:"hourly_limit"`
	LastResetHour time.Time          `bson:"last_reset_hour" json:"last_reset_hour"`
	LastResetDay  time.Time          `bson:"last_reset_day" json:"last_reset_day"`
	CreatedAt     time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at" json:"updated_at"`
}

// IPReputation IP信誉监控
type IPReputation struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	IP              string             `bson:"ip" json:"ip"`
	ReputationScore float64            `bson:"reputation_score" json:"reputation_score"`
	SuccessRate     float64            `bson:"success_rate" json:"success_rate"`
	TotalSent       int64              `bson:"total_sent" json:"total_sent"`
	TotalFailed     int64              `bson:"total_failed" json:"total_failed"`
	LastChecked     time.Time          `bson:"last_checked" json:"last_checked"`
	Status          string             `bson:"status" json:"status"` // good, warning, blocked
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}
