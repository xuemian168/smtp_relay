package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MailLog 邮件发送日志结构
type MailLog struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID  `bson:"user_id" json:"user_id"`
	CredentialID *primitive.ObjectID `bson:"credential_id,omitempty" json:"credential_id,omitempty"`
	MessageID    string              `bson:"message_id" json:"message_id"`
	From         string              `bson:"from" json:"from"`
	To           []string            `bson:"to" json:"to"`
	Subject      string              `bson:"subject" json:"subject"`
	Size         int64               `bson:"size" json:"size"`
	Status       string              `bson:"status" json:"status"` // queued, sending, sent, failed
	Attempts     int                 `bson:"attempts" json:"attempts"`
	LastAttempt  time.Time           `bson:"last_attempt" json:"last_attempt"`
	ErrorMessage string              `bson:"error_message,omitempty" json:"error_message,omitempty"`
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
	CompletedAt  *time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	RelayIP      string              `bson:"relay_ip" json:"relay_ip"`
}

// SMTPConfig SMTP服务器配置
type SMTPConfig struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name     string             `bson:"name" json:"name"`
	Host     string             `bson:"host" json:"host"`
	Port     int                `bson:"port" json:"port"`
	Username string             `bson:"username" json:"username"`
	Password string             `bson:"password" json:"-"`
	TLS      bool               `bson:"tls" json:"tls"`
	Active   bool               `bson:"active" json:"active"`
	Priority int                `bson:"priority" json:"priority"`
}
