package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"smtp-relay/internal/auth"
	"smtp-relay/internal/database"
	"smtp-relay/internal/models"
	"smtp-relay/internal/queue"
	"smtp-relay/internal/services"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"
	"github.com/sirupsen/logrus"
)

// Server SMTP服务器结构
type Server struct {
	config            *Config
	db                *database.MongoDB
	logger            *logrus.Logger
	auth              *auth.Service
	queue             *queue.Service
	credentialService *services.SMTPCredentialService
	server            *smtp.Server
}

// Config SMTP服务器配置
type Config struct {
	Host       string
	Port25     int
	Port587    int
	Port465    int
	Domain     string
	TLSCert    string
	TLSKey     string
	MaxMsgSize int64
}

// NewServer 创建SMTP服务器
func NewServer(config *Config, db *database.MongoDB, logger *logrus.Logger, auth *auth.Service, queue *queue.Service, credentialService *services.SMTPCredentialService) *Server {
	return &Server{
		config:            config,
		db:                db,
		logger:            logger,
		auth:              auth,
		queue:             queue,
		credentialService: credentialService,
	}
}

// Start 启动SMTP服务器
func (s *Server) Start() error {
	// 创建SMTP服务器
	s.server = smtp.NewServer(&Backend{
		server: s,
	})

	s.server.Addr = fmt.Sprintf("%s:%d", s.config.Host, s.config.Port25)
	s.server.Domain = s.config.Domain
	s.server.MaxMessageBytes = s.config.MaxMsgSize
	s.server.MaxRecipients = 100
	s.server.AllowInsecureAuth = true // 允许不安全认证（开发环境）
	s.server.EnableSMTPUTF8 = true

	// 配置TLS
	if s.config.TLSCert != "" && s.config.TLSKey != "" {
		cert, err := tls.LoadX509KeyPair(s.config.TLSCert, s.config.TLSKey)
		if err != nil {
			return fmt.Errorf("加载TLS证书失败: %w", err)
		}

		s.server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	s.logger.WithFields(logrus.Fields{
		"addr":    s.server.Addr,
		"domain":  s.server.Domain,
		"max_msg": s.config.MaxMsgSize,
	}).Info("启动SMTP服务器")

	// 启动多个端口监听
	go s.startPort25()
	go s.startPort587()
	go s.startPort465()

	return nil
}

// startPort25 启动端口25（标准SMTP）
func (s *Server) startPort25() {
	s.logger.WithField("port", s.config.Port25).Info("启动SMTP端口25")

	// 确保端口25也支持认证（用于测试）
	s.server.AllowInsecureAuth = true

	if err := s.server.ListenAndServe(); err != nil {
		s.logger.WithError(err).Error("SMTP端口25启动失败")
	}
}

// startPort587 启动端口587（SMTP提交）
func (s *Server) startPort587() {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port587)
	s.logger.WithField("port", s.config.Port587).Info("启动SMTP端口587")

	server587 := *s.server
	server587.Addr = addr
	server587.AllowInsecureAuth = true // 允许不安全认证

	if err := server587.ListenAndServe(); err != nil {
		s.logger.WithError(err).Error("SMTP端口587启动失败")
	}
}

// startPort465 启动端口465（SMTPS）
func (s *Server) startPort465() {
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port465)
	s.logger.WithField("port", s.config.Port465).Info("启动SMTP端口465")

	server465 := *s.server
	server465.Addr = addr
	server465.AllowInsecureAuth = true // 允许不安全认证

	// 如果没有TLS配置，使用普通监听而不是TLS监听
	if s.config.TLSCert == "" || s.config.TLSKey == "" {
		s.logger.WithField("port", s.config.Port465).Warn("端口465无TLS证书，使用普通监听")
		// 移除TLS配置以避免错误
		server465.TLSConfig = nil
		if err := server465.ListenAndServe(); err != nil {
			s.logger.WithError(err).Error("SMTP端口465启动失败")
		}
	} else {
		if err := server465.ListenAndServeTLS(); err != nil {
			s.logger.WithError(err).Error("SMTP端口465启动失败")
		}
	}
}

// Stop 停止SMTP服务器
func (s *Server) Stop() error {
	if s.server != nil {
		s.logger.Info("停止SMTP服务器")
		return s.server.Close()
	}
	return nil
}

// Backend SMTP后端实现
type Backend struct {
	server *Server
}

// NewSession 创建新的SMTP会话
func (b *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &Session{
		server: b.server,
		conn:   c,
		logger: b.server.logger.WithField("remote", c.Conn().RemoteAddr()),
	}, nil
}

// Session SMTP会话实现
type Session struct {
	server     *Server
	conn       *smtp.Conn
	logger     *logrus.Entry
	user       *models.User
	credential *models.SMTPCredential
	from       string
	to         []string
}

// AuthMechanisms 返回支持的认证机制
func (s *Session) AuthMechanisms() []string {
	return []string{"PLAIN"}
}

// Auth 处理认证
func (s *Session) Auth(mech string) (sasl.Server, error) {
	switch mech {
	case "PLAIN":
		return sasl.NewPlainServer(func(identity, username, password string) error {
			return s.AuthPlain(username, password)
		}), nil
	default:
		return nil, fmt.Errorf("unsupported auth mechanism: %s", mech)
	}
}

// AuthPlain 处理PLAIN认证
func (s *Session) AuthPlain(username, password string) error {
	s.logger.WithField("username", username).Info("SMTP认证请求")

	// 使用新的多密钥对认证
	credential, err := s.server.credentialService.AuthenticateSMTP(username, password)
	if err != nil {
		s.logger.WithError(err).WithField("username", username).Warn("SMTP认证失败")
		return err
	}

	// 获取用户信息
	userCollection := s.server.db.GetCollection("users")
	var user models.User
	err = userCollection.FindOne(context.Background(), bson.M{"_id": credential.UserID, "status": "active"}).Decode(&user)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", credential.UserID.Hex()).Error("获取用户信息失败")
		return fmt.Errorf("用户信息获取失败")
	}

	s.user = &user
	s.credential = credential
	s.logger.WithFields(logrus.Fields{
		"user_id":         user.ID.Hex(),
		"credential_id":   credential.ID.Hex(),
		"credential_name": credential.Name,
	}).Info("SMTP认证成功")

	return nil
}

// Mail 处理MAIL FROM命令
func (s *Session) Mail(from string, opts *smtp.MailOptions) error {
	s.logger.WithField("from", from).Info("收到MAIL FROM命令")

	// 强制要求认证，防止滥用
	if s.user == nil || s.credential == nil {
		s.logger.WithField("from", from).Warn("未认证用户尝试发送邮件")
		return fmt.Errorf("认证失败：必须先通过SMTP认证才能发送邮件")
	}

	// 验证发件人地址（使用凭据级别的域名限制）
	if !s.isValidSender(from) {
		s.logger.WithField("from", from).Warn("无效的发件人地址")
		return fmt.Errorf("无效的发件人地址: %s", from)
	}

	// 记录认证用户的发送行为
	s.logger.WithFields(logrus.Fields{
		"user_id":         s.user.ID.Hex(),
		"credential_id":   s.credential.ID.Hex(),
		"credential_name": s.credential.Name,
		"from":            from,
	}).Info("已认证用户发送邮件")

	s.from = from
	return nil
}

// Rcpt 处理RCPT TO命令
func (s *Session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.logger.WithField("to", to).Info("收到RCPT TO命令")

	// 强制要求认证，防止滥用
	if s.user == nil || s.credential == nil {
		s.logger.WithField("to", to).Warn("未认证用户尝试添加收件人")
		return fmt.Errorf("认证失败：必须先通过SMTP认证才能添加收件人")
	}

	// 检查收件人数量限制（使用凭据级别的设置）
	maxRecipients := s.credential.Settings.MaxRecipients
	if maxRecipients == 0 {
		maxRecipients = 100 // 默认值
	}

	if len(s.to) >= maxRecipients {
		s.logger.WithFields(logrus.Fields{
			"user_id":       s.user.ID.Hex(),
			"credential_id": s.credential.ID.Hex(),
			"to_count":      len(s.to),
			"max_allowed":   maxRecipients,
		}).Warn("收件人数量超过限制")
		return fmt.Errorf("收件人数量超过限制（最多%d个）", maxRecipients)
	}

	s.to = append(s.to, to)
	s.logger.WithFields(logrus.Fields{
		"user_id":       s.user.ID.Hex(),
		"credential_id": s.credential.ID.Hex(),
		"to":            to,
		"to_count":      len(s.to),
	}).Info("已认证用户添加收件人")

	return nil
}

// Data 处理邮件数据
func (s *Session) Data(r io.Reader) error {
	s.logger.WithFields(logrus.Fields{
		"from":          s.from,
		"to_count":      len(s.to),
		"user_id":       s.user.ID.Hex(),
		"credential_id": s.credential.ID.Hex(),
	}).Info("接收邮件数据")

	// 读取邮件内容
	data, err := io.ReadAll(r)
	if err != nil {
		s.logger.WithError(err).Error("读取邮件数据失败")
		return err
	}

	// 检查邮件大小
	if int64(len(data)) > s.server.config.MaxMsgSize {
		return fmt.Errorf("邮件大小超过限制")
	}

	// 检查凭据级别的配额
	if err := s.checkCredentialQuota(); err != nil {
		s.logger.WithError(err).Warn("凭据配额检查失败")
		return err
	}

	// 创建MailLog记录
	mailLog := &models.MailLog{
		UserID:       s.user.ID,
		CredentialID: &s.credential.ID,
		MessageID:    s.generateMessageID(),
		From:         s.from,
		To:           s.to,
		Subject:      s.extractSubject(string(data)),
		Size:         int64(len(data)),
		Status:       "queued",
		Attempts:     0,
		CreatedAt:    time.Now(),
		RelayIP:      s.getServerIP(),
	}

	// 将邮件加入队列
	if err := s.server.queue.EnqueueMail(mailLog, data); err != nil {
		s.logger.WithError(err).Error("邮件入队失败")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"message_id":    mailLog.MessageID,
		"credential_id": s.credential.ID.Hex(),
	}).Info("邮件已加入发送队列")
	return nil
}

// Reset 重置会话
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
}

// Logout 退出会话
func (s *Session) Logout() error {
	s.logger.Info("SMTP会话结束")
	return nil
}

// 辅助方法

// isValidSender 验证发件人地址（使用凭据级别的域名限制）
func (s *Session) isValidSender(from string) bool {
	// 检查是否在凭据允许的域名列表中
	if len(s.credential.Settings.AllowedDomains) > 0 {
		domain := strings.Split(from, "@")[1]
		for _, allowedDomain := range s.credential.Settings.AllowedDomains {
			if domain == allowedDomain {
				return true
			}
		}
		return false
	}
	return true
}

// generateMessageID 生成邮件ID
func (s *Session) generateMessageID() string {
	return fmt.Sprintf("%d-%s@%s", time.Now().Unix(), s.user.ID.Hex(), s.server.config.Domain)
}

// extractSubject 提取邮件主题
func (s *Session) extractSubject(data string) string {
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.ToLower(line), "subject:") {
			return strings.TrimSpace(line[8:])
		}
	}
	return ""
}

// getServerIP 获取服务器IP
func (s *Session) getServerIP() string {
	if conn, ok := s.conn.Conn().(*net.TCPConn); ok {
		if addr, ok := conn.LocalAddr().(*net.TCPAddr); ok {
			return addr.IP.String()
		}
	}
	return "unknown"
}

// checkCredentialQuota 检查凭据级别的配额
func (s *Session) checkCredentialQuota() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 检查日配额
	if s.credential.Settings.DailyQuota > 0 {
		dailyCount, err := s.getDailyMailCount(ctx)
		if err != nil {
			return err
		}
		if dailyCount >= int64(s.credential.Settings.DailyQuota) {
			return fmt.Errorf("凭据日配额已用完（%d/%d）", dailyCount, s.credential.Settings.DailyQuota)
		}
	}

	// 检查小时配额
	if s.credential.Settings.HourlyQuota > 0 {
		hourlyCount, err := s.getHourlyMailCount(ctx)
		if err != nil {
			return err
		}
		if hourlyCount >= int64(s.credential.Settings.HourlyQuota) {
			return fmt.Errorf("凭据小时配额已用完（%d/%d）", hourlyCount, s.credential.Settings.HourlyQuota)
		}
	}

	return nil
}

// getDailyMailCount 获取凭据今日发送邮件数量
func (s *Session) getDailyMailCount(ctx context.Context) (int64, error) {
	collection := s.server.db.GetCollection("mail_logs")

	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	filter := bson.M{
		"credential_id": s.credential.ID,
		"created_at": bson.M{
			"$gte": today,
			"$lt":  tomorrow,
		},
	}

	return collection.CountDocuments(ctx, filter)
}

// getHourlyMailCount 获取凭据本小时发送邮件数量
func (s *Session) getHourlyMailCount(ctx context.Context) (int64, error) {
	collection := s.server.db.GetCollection("mail_logs")

	thisHour := time.Now().Truncate(time.Hour)
	nextHour := thisHour.Add(time.Hour)

	filter := bson.M{
		"credential_id": s.credential.ID,
		"created_at": bson.M{
			"$gte": thisHour,
			"$lt":  nextHour,
		},
	}

	return collection.CountDocuments(ctx, filter)
}
