package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	_ "smtp-relay/docs"
	"smtp-relay/internal/auth"
	"smtp-relay/internal/database"
	"smtp-relay/internal/models"
	"smtp-relay/internal/services"
)

// 请求结构体定义

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"testuser"`
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"password123"`
}

// CreateCredentialRequest 创建SMTP凭据请求
type CreateCredentialRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=50" example:"My SMTP Credential"`
	Description string `json:"description" binding:"max=200" example:"用于发送营销邮件的SMTP凭据"`
}

// UpdateCredentialRequest 更新SMTP凭据请求
type UpdateCredentialRequest struct {
	Name        string                         `json:"name" binding:"required,min=1,max=50" example:"Updated SMTP Credential"`
	Description string                         `json:"description" binding:"max=200" example:"更新后的SMTP凭据描述"`
	Settings    *models.SMTPCredentialSettings `json:"settings"`
}

// UpdateUserInfoRequest 更新用户信息请求
type UpdateUserInfoRequest struct {
	Username string               `json:"username" binding:"omitempty,min=3,max=50" example:"newusername"`
	Settings *models.UserSettings `json:"settings"`
}

// GetMailLogsRequest 获取MailLog请求参数
type GetMailLogsRequest struct {
	Page     int    `form:"page,default=1"`
	PageSize int    `form:"page_size,default=20"`
	Status   string `form:"status"`
	From     string `form:"from"`
	To       string `form:"to"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

// GetRecentMailLogsRequest 获取近期MailLog请求参数
type GetRecentMailLogsRequest struct {
	Days     int    `form:"days,default=30"`      // 最近N天，默认30天
	Status   string `form:"status"`               // 状态过滤
	PageSize int    `form:"page_size,default=50"` // 每页记录数，默认50
	Page     int    `form:"page,default=1"`       // 页码
}

// 响应结构体定义

// APIResponse 标准API响应
type APIResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message,omitempty" example:"操作成功"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:"错误信息"`
}

// UserInfo 用户信息
type UserInfo struct {
	ID       string `json:"id" example:"507f1f77bcf86cd799439011"`
	Username string `json:"username" example:"testuser"`
	Email    string `json:"email" example:"user@example.com"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Success bool     `json:"success" example:"true"`
	Token   string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    UserInfo `json:"user"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	Success bool     `json:"success" example:"true"`
	Token   string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User    UserInfo `json:"user"`
}

// UserInfoResponse 用户信息响应
type UserInfoResponse struct {
	ID        string               `json:"id" example:"507f1f77bcf86cd799439011"`
	Username  string               `json:"username" example:"testuser"`
	Email     string               `json:"email" example:"user@example.com"`
	Status    string               `json:"status" example:"active"`
	Settings  *models.UserSettings `json:"settings"`
	CreatedAt time.Time            `json:"created_at" example:"2023-01-01T00:00:00Z"`
}

// CredentialResponse SMTP凭据响应
type CredentialResponse struct {
	Success bool                   `json:"success" example:"true"`
	Data    *models.SMTPCredential `json:"data"`
}

// CredentialListResponse SMTP凭据列表响应
type CredentialListResponse struct {
	Success bool                     `json:"success" example:"true"`
	Data    []*models.SMTPCredential `json:"data"`
}

// CreateCredentialResponse 创建SMTP凭据响应
type CreateCredentialResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		Credential *models.SMTPCredential `json:"credential"`
		Password   string                 `json:"password" example:"generated_password"`
	} `json:"data"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	Success  bool   `json:"success" example:"true"`
	Message  string `json:"message" example:"密码重置成功"`
	Password string `json:"password" example:"new_generated_password"`
}

// MailLogResponse MailLog响应
type MailLogResponse struct {
	Success bool            `json:"success" example:"true"`
	Data    *models.MailLog `json:"data"`
}

// MailLogListResponse MailLog列表响应
type MailLogListResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		MailLogs []*models.MailLog `json:"mail_logs"`
		Total    int64             `json:"total" example:"100"`
		Page     int               `json:"page" example:"1"`
		PageSize int               `json:"page_size" example:"20"`
		Pages    int64             `json:"pages" example:"5"`
	} `json:"data"`
}

// RecentMailLogsResponse 近期MailLog响应
type RecentMailLogsResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		MailLogs   []*models.MailLog `json:"mail_logs"`
		Total      int64             `json:"total" example:"100"`
		Page       int               `json:"page" example:"1"`
		PageSize   int               `json:"page_size" example:"50"`
		Pages      int64             `json:"pages" example:"2"`
		Days       int               `json:"days" example:"30"`
		Statistics map[string]int64  `json:"statistics"`
	} `json:"data"`
}

// StatsResponse 统计信息响应
type StatsResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		MailStats       map[string]interface{}   `json:"mail_stats"`
		CredentialStats []map[string]interface{} `json:"credential_stats"`
		CredentialCount int                      `json:"credential_count" example:"3"`
	} `json:"data"`
}

// QuotaStatsResponse 配额统计响应
type QuotaStatsResponse struct {
	Success bool `json:"success" example:"true"`
	Data    struct {
		UserSettings struct {
			DailyQuota  int `json:"daily_quota" example:"1000"`
			HourlyQuota int `json:"hourly_quota" example:"100"`
		} `json:"user_settings"`
		CredentialQuotas []map[string]interface{} `json:"credential_quotas"`
	} `json:"data"`
}

// Server API服务器结构
type Server struct {
	config            *Config
	db                *database.MongoDB
	logger            *logrus.Logger
	authService       *auth.Service
	credentialService *services.SMTPCredentialService
	mailLogService    *services.MailLogService
	router            *gin.Engine
	server            *http.Server
}

// Config API服务器配置
type Config struct {
	Port      string
	SecretKey string
}

// NewServer 创建API服务器
func NewServer(config *Config, db *database.MongoDB, logger *logrus.Logger, authService *auth.Service, credentialService *services.SMTPCredentialService, mailLogService *services.MailLogService) *Server {
	return &Server{
		config:            config,
		db:                db,
		logger:            logger,
		authService:       authService,
		credentialService: credentialService,
		mailLogService:    mailLogService,
	}
}

// Start 启动API服务器
func (s *Server) Start() error {
	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	s.router = gin.New()

	// 添加中间件
	s.router.Use(gin.Recovery())
	s.router.Use(s.loggerMiddleware())
	s.router.Use(s.corsMiddleware())

	// 设置路由
	s.setupRoutes()

	// 创建HTTP服务器
	s.server = &http.Server{
		Addr:         ":" + s.config.Port,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	s.logger.WithField("port", s.config.Port).Info("启动API服务器")

	// 启动服务器
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("启动API服务器失败: %w", err)
	}

	return nil
}

// Stop 停止API服务器
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.logger.Info("停止API服务器")
	return s.server.Shutdown(ctx)
}

// setupRoutes 设置路由
func (s *Server) setupRoutes() {
	// 健康检查
	s.router.GET("/health", s.healthCheck)

	// Swagger文档
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API版本1
	v1 := s.router.Group("/api/v1")
	{
		// 用户认证
		auth := v1.Group("/auth")
		{
			auth.POST("/login", s.login)
			auth.POST("/register", s.register)
		}

		// 需要认证的路由
		authenticated := v1.Group("/")
		authenticated.Use(s.authMiddleware())
		{
			// 用户信息
			authenticated.GET("/user", s.getUserInfo)
			authenticated.PUT("/user", s.updateUserInfo)

			// SMTP凭据管理
			credentials := authenticated.Group("/credentials")
			{
				credentials.GET("", s.listCredentials)
				credentials.POST("", s.createCredential)
				credentials.GET("/:id", s.getCredential)
				credentials.PUT("/:id", s.updateCredential)
				credentials.DELETE("/:id", s.deleteCredential)
				credentials.POST("/:id/reset-password", s.resetCredentialPassword)
			}

			// MailLog
			logs := authenticated.Group("/logs")
			{
				logs.GET("", s.getMailLogs)
				logs.GET("/recent", s.getRecentMailLogs)
				logs.GET("/:id", s.getMailLog)
			}

			// 统计信息
			stats := authenticated.Group("/stats")
			{
				stats.GET("", s.getStats)
				stats.GET("/quota", s.getQuotaStats)
			}
		}
	}
}

// 中间件

// loggerMiddleware 日志中间件
func (s *Server) loggerMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		s.logger.WithFields(logrus.Fields{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency,
			"ip":         param.ClientIP,
			"user_agent": param.Request.UserAgent(),
		}).Info("API请求")
		return ""
	})
}

// corsMiddleware CORS中间件
func (s *Server) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		fmt.Println("[CORS DEBUG] Received Origin:", origin)
		allowedOrigins := map[string]bool{
			"http://0.0.0.0:3000":   true,
			"http://localhost:3000": true,
			"http://127.0.0.1:3000": true,
		}
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		reqHeaders := c.GetHeader("Access-Control-Request-Headers")
		fmt.Println("[CORS DEBUG] Access-Control-Request-Headers:", reqHeaders)
		if reqHeaders != "" {
			c.Header("Access-Control-Allow-Headers", reqHeaders)
		} else {
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, Accept, Accept-Language, X-Requested-With")
		}
		c.Header("Content-Type", "application/json; charset=utf-8")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// authMiddleware 认证中间件
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, gin.H{"error": "缺少认证令牌"})
			c.Abort()
			return
		}

		// 移除Bearer前缀
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// 验证JWT令牌
		userID, err := s.authService.ValidateJWT(token)
		if err != nil {
			c.JSON(401, gin.H{"error": "无效的认证令牌"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}

// 处理函数

// healthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务健康状态
// @tags system
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "服务正常"
// @Router /health [get]
func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status": "ok",
		"time":   time.Now().Unix(),
	})
}

// login 用户登录
// @Summary 用户登录
// @Description 用户登录获取JWT令牌
// @tags auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "登录信息"
// @Success 200 {object} LoginResponse "登录成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 401 {object} APIResponse "认证失败"
// @Router /api/v1/auth/login [post]
func (s *Server) login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 支持邮箱或用户名登录
	user, err := s.authService.AuthenticateUser(req.Username, req.Password)
	if err != nil {
		// 如果按邮箱查找失败，尝试按用户名查找
		user, err = s.authService.AuthenticateUserByUsername(req.Username, req.Password)
		if err != nil {
			c.JSON(401, gin.H{"error": "账号或密码错误"})
			return
		}
	}

	// 生成JWT令牌
	token, err := s.authService.GenerateJWT(user.ID)
	if err != nil {
		s.logger.WithError(err).Error("生成JWT令牌失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(200, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// register 用户注册
// @Summary 用户注册
// @Description 注册新用户账户
// @tags auth
// @Accept json
// @Produce json
// @Param body body RegisterRequest true "注册信息"
// @Success 201 {object} RegisterResponse "注册成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Router /api/v1/auth/register [post]
func (s *Server) register(c *gin.Context) {
	var req RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 创建用户
	user, err := s.authService.CreateUser(req.Username, req.Email, req.Password)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// 生成JWT令牌
	token, err := s.authService.GenerateJWT(user.ID)
	if err != nil {
		s.logger.WithError(err).Error("生成JWT令牌失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(201, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
		},
	})
}

// getUserInfo 获取用户信息
// @Summary 获取用户信息
// @Description 获取当前用户的详细信息
// @tags usermgmt
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserInfoResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/user [get]
func (s *Server) getUserInfo(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := s.authService.GetUserByID(userID)
	if err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(200, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"status":     user.Status,
		"settings":   user.Settings,
		"created_at": user.CreatedAt,
	})
}

// updateUserInfo 更新用户信息
// @Summary 更新用户信息
// @Description 更新当前用户的信息
// @tags usermgmt
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body UpdateUserInfoRequest true "更新信息"
// @Success 200 {object} APIResponse "更新成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Router /api/v1/user [put]
func (s *Server) updateUserInfo(c *gin.Context) {
	var req UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取当前用户信息
	user, err := s.authService.GetUserByID(userID.Hex())
	if err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	// 构建更新数据
	updateData := bson.M{
		"updated_at": time.Now(),
	}

	// 更新用户名（如果提供）
	if req.Username != "" && req.Username != user.Username {
		// 检查用户名是否已存在
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		userCollection := s.db.GetCollection("users")
		count, err := userCollection.CountDocuments(ctx, bson.M{"username": req.Username})
		if err != nil {
			s.logger.WithError(err).Error("检查用户名重复失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
			return
		}
		if count > 0 {
			c.JSON(400, gin.H{"error": "用户名已存在"})
			return
		}

		updateData["username"] = req.Username
	}

	// 更新设置（如果提供）
	if req.Settings != nil {
		// 验证配额设置
		if req.Settings.DailyQuota < 0 || req.Settings.DailyQuota > 10000 {
			c.JSON(400, gin.H{"error": "日配额必须在0-10000之间"})
			return
		}
		if req.Settings.HourlyQuota < 0 || req.Settings.HourlyQuota > 1000 {
			c.JSON(400, gin.H{"error": "小时配额必须在0-1000之间"})
			return
		}

		updateData["settings"] = req.Settings
	}

	// 执行更新
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userCollection := s.db.GetCollection("users")
	filter := bson.M{"_id": userID}
	update := bson.M{"$set": updateData}

	result, err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("更新用户信息失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	s.logger.WithField("user_id", userID.Hex()).Info("用户信息更新成功")

	c.JSON(200, gin.H{
		"success": true,
		"message": "用户信息更新成功",
	})
}

// 辅助函数
func (s *Server) getUserObjectID(c *gin.Context) (primitive.ObjectID, error) {
	userIDStr := c.GetString("user_id")
	return primitive.ObjectIDFromHex(userIDStr)
}

func (s *Server) getCredentialID(c *gin.Context) (primitive.ObjectID, error) {
	credentialIDStr := c.Param("id")
	return primitive.ObjectIDFromHex(credentialIDStr)
}

// listCredentials 列出SMTP凭据
// @Summary 获取SMTP凭据列表
// @Description 获取当前用户的所有SMTP凭据
// @Tags SMTP Credentials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} CredentialListResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/credentials [get]
func (s *Server) listCredentials(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 调用服务层获取凭据列表
	credentials, err := s.credentialService.ListCredentials(userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("获取SMTP凭据列表失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    credentials,
	})
}

// createCredential 创建SMTP凭据
// @Summary 创建SMTP凭据
// @Description 为当前用户创建新的SMTP凭据
// @Tags SMTP Credentials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateCredentialRequest true "凭据信息"
// @Success 201 {object} CreateCredentialResponse "创建成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/credentials [post]
func (s *Server) createCredential(c *gin.Context) {
	var req CreateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 调用服务层创建凭据
	credential, password, err := s.credentialService.CreateCredential(userID, req.Name, req.Description)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("创建SMTP凭据失败")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data": gin.H{
			"credential": credential,
			"password":   password,
		},
	})
}

// getCredential 获取SMTP凭据
// @Summary 获取单个SMTP凭据
// @Description 获取指定ID的SMTP凭据详情
// @Tags SMTP Credentials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "凭据ID"
// @Success 200 {object} CredentialResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "凭据不存在"
// @Router /api/v1/credentials/{id} [get]
func (s *Server) getCredential(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取凭据ID
	credentialID, err := s.getCredentialID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的凭据ID"})
		return
	}

	// 调用服务层获取凭据
	credential, err := s.credentialService.GetCredential(userID, credentialID)
	if err != nil {
		if err.Error() == "SMTP凭据不存在" {
			c.JSON(404, gin.H{"error": "SMTP凭据不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":       userID.Hex(),
				"credential_id": credentialID.Hex(),
			}).Error("获取SMTP凭据失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    credential,
	})
}

// updateCredential 更新SMTP凭据
// @Summary 更新SMTP凭据
// @Description 更新指定ID的SMTP凭据信息
// @Tags SMTP Credentials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "凭据ID"
// @Param body body UpdateCredentialRequest true "更新信息"
// @Success 200 {object} APIResponse "更新成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "凭据不存在"
// @Router /api/v1/credentials/{id} [put]
func (s *Server) updateCredential(c *gin.Context) {
	var req UpdateCredentialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取凭据ID
	credentialID, err := s.getCredentialID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的凭据ID"})
		return
	}

	// 如果没有提供settings，使用默认值
	settings := models.SMTPCredentialSettings{
		DailyQuota:    1000,
		HourlyQuota:   100,
		MaxRecipients: 100,
	}
	if req.Settings != nil {
		settings = *req.Settings
	}

	// 调用服务层更新凭据
	err = s.credentialService.UpdateCredential(userID, credentialID, req.Name, req.Description, settings)
	if err != nil {
		if err.Error() == "SMTP凭据不存在" {
			c.JSON(404, gin.H{"error": "SMTP凭据不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":       userID.Hex(),
				"credential_id": credentialID.Hex(),
			}).Error("更新SMTP凭据失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "SMTP凭据更新成功",
	})
}

// deleteCredential 删除SMTP凭据
// @Summary 删除SMTP凭据
// @Description 删除指定ID的SMTP凭据
// @Tags SMTP Credentials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "凭据ID"
// @Success 200 {object} APIResponse "删除成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "凭据不存在"
// @Router /api/v1/credentials/{id} [delete]
func (s *Server) deleteCredential(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取凭据ID
	credentialID, err := s.getCredentialID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的凭据ID"})
		return
	}

	// 调用服务层删除凭据
	err = s.credentialService.DeleteCredential(userID, credentialID)
	if err != nil {
		if err.Error() == "SMTP凭据不存在" {
			c.JSON(404, gin.H{"error": "SMTP凭据不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":       userID.Hex(),
				"credential_id": credentialID.Hex(),
			}).Error("删除SMTP凭据失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "SMTP凭据删除成功",
	})
}

// resetCredentialPassword 重置凭据密码
// @Summary 重置SMTP凭据密码
// @Description 重置指定ID的SMTP凭据密码
// @Tags SMTP Credentials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "凭据ID"
// @Success 200 {object} ResetPasswordResponse "重置成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "凭据不存在"
// @Router /api/v1/credentials/{id}/reset-password [post]
func (s *Server) resetCredentialPassword(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取凭据ID
	credentialID, err := s.getCredentialID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的凭据ID"})
		return
	}

	// 调用服务层重置密码
	newPassword, err := s.credentialService.ResetPassword(userID, credentialID)
	if err != nil {
		if err.Error() == "SMTP凭据不存在" {
			c.JSON(404, gin.H{"error": "SMTP凭据不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":       userID.Hex(),
				"credential_id": credentialID.Hex(),
			}).Error("重置SMTP凭据密码失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success":  true,
		"message":  "密码重置成功",
		"password": newPassword,
	})
}

// getMailLogs 获取MailLog
// @Summary 获取MailLog
// @Description 获取当前用户的邮件发送日志，支持分页和筛选
// @Tags MailLog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param status query string false "邮件状态" Enums(queued,sending,sent,failed)
// @Param from query string false "发件人筛选"
// @Param to query string false "收件人筛选"
// @Param date_from query string false "开始日期" format(date)
// @Param date_to query string false "结束日期" format(date)
// @Success 200 {object} MailLogListResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/logs [get]
func (s *Server) getMailLogs(c *gin.Context) {
	var req GetMailLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 调用服务层获取MailLog
	mailLogs, total, err := s.mailLogService.GetMailLogsByUser(
		userID, req.Page, req.PageSize, req.Status, req.From, req.To, req.DateFrom, req.DateTo,
	)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("获取MailLog失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"mail_logs": mailLogs,
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
			"pages":     (total + int64(req.PageSize) - 1) / int64(req.PageSize),
		},
	})
}

// getMailLog 获取单个MailLog
// @Summary 获取单个MailLog
// @Description 获取指定ID的MailLog详情
// @Tags MailLog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "MailLogID"
// @Success 200 {object} MailLogResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "MailLog不存在"
// @Router /api/v1/logs/{id} [get]
func (s *Server) getMailLog(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取MailLogID
	mailLogIDStr := c.Param("id")
	mailLogID, err := primitive.ObjectIDFromHex(mailLogIDStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的MailLogID"})
		return
	}

	// 调用服务层获取MailLog
	mailLog, err := s.mailLogService.GetMailLogByID(userID, mailLogID)
	if err != nil {
		s.logger.WithError(err).WithFields(logrus.Fields{
			"user_id":     userID.Hex(),
			"mail_log_id": mailLogID.Hex(),
		}).Error("获取MailLog失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	if mailLog == nil {
		c.JSON(404, gin.H{"error": "MailLog不存在"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    mailLog,
	})
}

// getRecentMailLogs 获取近期MailLog
// @Summary 获取近期MailLog
// @Description 获取用户近期发信历史，支持天数筛选和状态筛选，包含统计信息
// @Tags MailLog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param days query int false "最近N天" default(30) minimum(1) maximum(365)
// @Param status query string false "邮件状态" Enums(queued,sending,sent,failed)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(50) minimum(1) maximum(100)
// @Success 200 {object} RecentMailLogsResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/logs/recent [get]
func (s *Server) getRecentMailLogs(c *gin.Context) {
	var req GetRecentMailLogsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(400, gin.H{"error": "请求参数错误"})
		return
	}

	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 参数验证
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 50
	}
	if req.Days < 1 || req.Days > 365 {
		req.Days = 30
	}

	// 调用服务层获取近期MailLog
	mailLogs, total, statistics, err := s.mailLogService.GetRecentMailLogsByUser(
		userID, req.Days, req.Page, req.PageSize, req.Status,
	)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("获取近期MailLog失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"mail_logs":  mailLogs,
			"total":      total,
			"page":       req.Page,
			"page_size":  req.PageSize,
			"pages":      (total + int64(req.PageSize) - 1) / int64(req.PageSize),
			"days":       req.Days,
			"statistics": statistics,
		},
	})
}

// getStats 获取统计信息
// @Summary 获取统计信息
// @Description 获取用户邮件统计信息和凭据统计信息
// @tags status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} StatsResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/stats [get]
func (s *Server) getStats(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取邮件统计信息
	mailStats, err := s.mailLogService.GetUserMailStats(userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("获取邮件统计信息失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	// 获取凭据统计信息
	credentials, err := s.credentialService.ListCredentials(userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("获取凭据列表失败")
		// 继续执行，不影响主要统计信息
	}

	credentialStats := make([]map[string]interface{}, 0)
	for _, credential := range credentials {
		credStats, err := s.mailLogService.GetCredentialMailStats(userID, credential.ID)
		if err != nil {
			continue
		}
		credentialStats = append(credentialStats, map[string]interface{}{
			"credential_id":   credential.ID,
			"credential_name": credential.Name,
			"today_total":     credStats["today_total"],
			"hour_total":      credStats["hour_total"],
		})
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"mail_stats":       mailStats,
			"credential_stats": credentialStats,
			"credential_count": len(credentials),
		},
	})
}

// getQuotaStats 获取配额统计
// @Summary 获取配额统计
// @Description 获取用户邮件配额使用情况统计
// @tags status
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} QuotaStatsResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/stats/quota [get]
func (s *Server) getQuotaStats(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取用户信息
	user, err := s.authService.GetUserByID(userID.Hex())
	if err != nil {
		c.JSON(404, gin.H{"error": "用户不存在"})
		return
	}

	// 获取凭据列表和配额信息
	credentials, err := s.credentialService.ListCredentials(userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("获取凭据列表失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	quotaStats := make([]map[string]interface{}, 0)
	for _, credential := range credentials {
		credStats, err := s.mailLogService.GetCredentialMailStats(userID, credential.ID)
		if err != nil {
			continue
		}

		todayUsed := int64(0)
		hourUsed := int64(0)
		if credStats["today_total"] != nil {
			todayUsed = credStats["today_total"].(int64)
		}
		if credStats["hour_total"] != nil {
			hourUsed = credStats["hour_total"].(int64)
		}

		quotaStats = append(quotaStats, map[string]interface{}{
			"credential_id":    credential.ID,
			"credential_name":  credential.Name,
			"daily_quota":      credential.Settings.DailyQuota,
			"daily_used":       todayUsed,
			"daily_remaining":  int64(credential.Settings.DailyQuota) - todayUsed,
			"hourly_quota":     credential.Settings.HourlyQuota,
			"hourly_used":      hourUsed,
			"hourly_remaining": int64(credential.Settings.HourlyQuota) - hourUsed,
		})
	}

	c.JSON(200, gin.H{
		"success": true,
		"data": gin.H{
			"user_settings": gin.H{
				"daily_quota":  user.Settings.DailyQuota,
				"hourly_quota": user.Settings.HourlyQuota,
			},
			"credential_quotas": quotaStats,
		},
	})
}
