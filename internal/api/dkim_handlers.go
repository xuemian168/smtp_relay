package api

import (
	"smtp-relay/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// DKIM相关请求结构体

// CreateDKIMKeyRequest 创建DKIM密钥对请求
type CreateDKIMKeyRequest struct {
	Domain   string `json:"domain" binding:"required" example:"example.com"`
	Selector string `json:"selector" binding:"required" example:"default"`
	KeySize  int    `json:"key_size" binding:"omitempty" example:"2048"`
}

// DKIM相关响应结构体

// DKIMKeyPairResponse DKIM密钥对响应
type DKIMKeyPairResponse struct {
	Success bool                `json:"success" example:"true"`
	Data    *models.DKIMKeyPair `json:"data"`
}

// DKIMKeyPairListResponse DKIM密钥对列表响应
type DKIMKeyPairListResponse struct {
	Success bool                  `json:"success" example:"true"`
	Data    []*models.DKIMKeyPair `json:"data"`
}

// DNSRecordResponse DNS记录响应
type DNSRecordResponse struct {
	Success bool              `json:"success" example:"true"`
	Data    *models.DNSRecord `json:"data"`
}

// DKIMValidationResponse DKIM验证响应
type DKIMValidationResponse struct {
	Success bool                         `json:"success" example:"true"`
	Data    *models.DKIMValidationResult `json:"data"`
}

// setupDKIMRoutes 设置DKIM相关路由
func (s *Server) setupDKIMRoutes(authenticated *gin.RouterGroup) {
	dkim := authenticated.Group("/dkim")
	{
		dkim.GET("/keys", s.listDKIMKeys)
		dkim.POST("/keys", s.createDKIMKey)
		dkim.GET("/keys/:id", s.getDKIMKey)
		dkim.DELETE("/keys/:id", s.deleteDKIMKey)
		dkim.POST("/keys/:id/rotate", s.rotateDKIMKey)
		dkim.GET("/keys/:id/dns", s.getDKIMDNSRecord)
		dkim.POST("/keys/:id/verify", s.verifyDKIMDNS)
	}
}

// listDKIMKeys 获取DKIM密钥对列表
// @Summary 获取DKIM密钥对列表
// @Description 获取当前用户的所有DKIM密钥对
// @Tags DKIM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} DKIMKeyPairListResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/dkim/keys [get]
func (s *Server) listDKIMKeys(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 调用服务层获取密钥对列表
	keyPairs, err := s.dkimService.ListKeyPairs(userID)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("获取DKIM密钥对列表失败")
		c.JSON(500, gin.H{"error": "服务器内部错误"})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    keyPairs,
	})
}

// createDKIMKey 创建DKIM密钥对
// @Summary 创建DKIM密钥对
// @Description 为指定域名创建新的DKIM密钥对
// @Tags DKIM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body CreateDKIMKeyRequest true "密钥对信息"
// @Success 201 {object} DKIMKeyPairResponse "创建成功"
// @Failure 400 {object} APIResponse "请求参数错误"
// @Failure 401 {object} APIResponse "未授权"
// @Router /api/v1/dkim/keys [post]
func (s *Server) createDKIMKey(c *gin.Context) {
	var req CreateDKIMKeyRequest
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

	// 设置默认密钥长度
	if req.KeySize == 0 {
		req.KeySize = 2048
	}

	// 调用服务层创建密钥对
	keyPair, err := s.dkimService.GenerateKeyPair(userID, req.Domain, req.Selector, req.KeySize)
	if err != nil {
		s.logger.WithError(err).WithField("user_id", userID.Hex()).Error("创建DKIM密钥对失败")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"success": true,
		"data":    keyPair,
	})
}

// getDKIMKey 获取DKIM密钥对
// @Summary 获取单个DKIM密钥对
// @Description 获取指定ID的DKIM密钥对详情
// @Tags DKIM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "密钥对ID"
// @Success 200 {object} DKIMKeyPairResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "密钥对不存在"
// @Router /api/v1/dkim/keys/{id} [get]
func (s *Server) getDKIMKey(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取密钥对ID
	keyPairID, err := s.getDKIMKeyID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的密钥对ID"})
		return
	}

	// 调用服务层获取密钥对
	keyPair, err := s.dkimService.GetKeyPair(userID, keyPairID)
	if err != nil {
		if err.Error() == "DKIM密钥对不存在" {
			c.JSON(404, gin.H{"error": "DKIM密钥对不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":     userID.Hex(),
				"key_pair_id": keyPairID.Hex(),
			}).Error("获取DKIM密钥对失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    keyPair,
	})
}

// deleteDKIMKey 删除DKIM密钥对
// @Summary 删除DKIM密钥对
// @Description 删除指定ID的DKIM密钥对
// @Tags DKIM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "密钥对ID"
// @Success 200 {object} APIResponse "删除成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "密钥对不存在"
// @Router /api/v1/dkim/keys/{id} [delete]
func (s *Server) deleteDKIMKey(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取密钥对ID
	keyPairID, err := s.getDKIMKeyID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的密钥对ID"})
		return
	}

	// 调用服务层删除密钥对
	err = s.dkimService.DeleteKeyPair(userID, keyPairID)
	if err != nil {
		if err.Error() == "DKIM密钥对不存在" {
			c.JSON(404, gin.H{"error": "DKIM密钥对不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":     userID.Hex(),
				"key_pair_id": keyPairID.Hex(),
			}).Error("删除DKIM密钥对失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "DKIM密钥对删除成功",
	})
}

// rotateDKIMKey 轮换DKIM密钥对
// @Summary 轮换DKIM密钥对
// @Description 为指定密钥对生成新的密钥，旧密钥标记为即将过期
// @Tags DKIM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "密钥对ID"
// @Success 200 {object} DKIMKeyPairResponse "轮换成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "密钥对不存在"
// @Router /api/v1/dkim/keys/{id}/rotate [post]
func (s *Server) rotateDKIMKey(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取密钥对ID
	keyPairID, err := s.getDKIMKeyID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的密钥对ID"})
		return
	}

	// 调用服务层轮换密钥对
	newKeyPair, err := s.dkimService.RotateKeyPair(userID, keyPairID)
	if err != nil {
		if err.Error() == "DKIM密钥对不存在" {
			c.JSON(404, gin.H{"error": "DKIM密钥对不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":     userID.Hex(),
				"key_pair_id": keyPairID.Hex(),
			}).Error("轮换DKIM密钥对失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "DKIM密钥对轮换成功",
		"data":    newKeyPair,
	})
}

// getDKIMDNSRecord 获取DKIM DNS记录
// @Summary 获取DKIM DNS记录
// @Description 获取指定密钥对的DNS TXT记录信息
// @Tags DKIM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "密钥对ID"
// @Success 200 {object} DNSRecordResponse "获取成功"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "密钥对不存在"
// @Router /api/v1/dkim/keys/{id}/dns [get]
func (s *Server) getDKIMDNSRecord(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取密钥对ID
	keyPairID, err := s.getDKIMKeyID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的密钥对ID"})
		return
	}

	// 调用服务层获取DNS记录
	dnsRecord, err := s.dkimService.GetDNSRecord(userID, keyPairID)
	if err != nil {
		if err.Error() == "DKIM密钥对不存在" {
			c.JSON(404, gin.H{"error": "DKIM密钥对不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":     userID.Hex(),
				"key_pair_id": keyPairID.Hex(),
			}).Error("获取DKIM DNS记录失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    dnsRecord,
	})
}

// verifyDKIMDNS 验证DKIM DNS记录
// @Summary 验证DKIM DNS记录
// @Description 验证指定密钥对的DNS记录是否正确配置
// @Tags DKIM
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "密钥对ID"
// @Success 200 {object} DKIMValidationResponse "验证完成"
// @Failure 401 {object} APIResponse "未授权"
// @Failure 404 {object} APIResponse "密钥对不存在"
// @Router /api/v1/dkim/keys/{id}/verify [post]
func (s *Server) verifyDKIMDNS(c *gin.Context) {
	// 获取用户ID
	userID, err := s.getUserObjectID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的用户ID"})
		return
	}

	// 获取密钥对ID
	keyPairID, err := s.getDKIMKeyID(c)
	if err != nil {
		c.JSON(400, gin.H{"error": "无效的密钥对ID"})
		return
	}

	// 调用服务层验证DNS记录
	result, err := s.dkimService.VerifyDNSRecord(userID, keyPairID)
	if err != nil {
		if err.Error() == "DKIM密钥对不存在" {
			c.JSON(404, gin.H{"error": "DKIM密钥对不存在"})
		} else {
			s.logger.WithError(err).WithFields(logrus.Fields{
				"user_id":     userID.Hex(),
				"key_pair_id": keyPairID.Hex(),
			}).Error("验证DKIM DNS记录失败")
			c.JSON(500, gin.H{"error": "服务器内部错误"})
		}
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    result,
	})
}

// 辅助函数

// getDKIMKeyID 从URL参数获取DKIM密钥对ID
func (s *Server) getDKIMKeyID(c *gin.Context) (primitive.ObjectID, error) {
	keyIDStr := c.Param("id")
	return primitive.ObjectIDFromHex(keyIDStr)
}
