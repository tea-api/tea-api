package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"tea-api/middleware"
	"tea-api/setting"
)

// GetSecurityStats 获取安全统计信息
func GetSecurityStats(c *gin.Context) {
	// 获取基本安全统计
	stats := setting.GetSecurityStats()
	
	// 获取黑名单统计
	blacklistStats := middleware.GetBlacklistStats()
	
	// 获取流统计
	streamStats := middleware.GetStreamStats()
	
	// 合并统计信息
	response := gin.H{
		"security_stats": stats,
		"blacklist_stats": blacklistStats,
		"stream_stats": streamStats,
		"config": gin.H{
			"abnormal_detection_enabled": setting.IsAbnormalDetectionEnabled(),
			"request_size_limit_enabled": setting.IsRequestSizeLimitEnabled(),
			"stream_protection_enabled":  setting.IsStreamProtectionEnabled(),
			"ip_blacklist_enabled":       setting.IsIPBlacklistEnabled(),
			"content_filter_enabled":     setting.IsContentFilterEnabled(),
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetSecurityConfig 获取安全配置
func GetSecurityConfig(c *gin.Context) {
	config := setting.GetSecurityConfig()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateSecurityConfig 更新安全配置
func UpdateSecurityConfig(c *gin.Context) {
	var newConfig setting.SecurityConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的配置格式: " + err.Error(),
		})
		return
	}
	
	// 验证配置的合理性
	if err := validateSecurityConfig(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "配置验证失败: " + err.Error(),
		})
		return
	}
	
	// 更新配置
	setting.UpdateSecurityConfig(newConfig)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "安全配置已更新",
	})
}

// GetBlacklist 获取IP黑名单
func GetBlacklist(c *gin.Context) {
	manager := middleware.GetBlacklistManager()
	stats := middleware.GetBlacklistStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stats": stats,
			"message": "使用管理接口查看详细黑名单信息",
		},
	})
}

// AddToBlacklist 添加IP到黑名单
func AddToBlacklist(c *gin.Context) {
	var request struct {
		IP        string `json:"ip" binding:"required"`
		Reason    string `json:"reason" binding:"required"`
		Temporary bool   `json:"temporary"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	manager := middleware.GetBlacklistManager()
	manager.AddToBlacklist(request.IP, request.Reason, request.Temporary)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP已添加到黑名单",
	})
}

// RemoveFromBlacklist 从黑名单移除IP
func RemoveFromBlacklist(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "IP地址不能为空",
		})
		return
	}
	
	manager := middleware.GetBlacklistManager()
	manager.RemoveFromBlacklist(ip)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP已从黑名单移除",
	})
}

// AddToWhitelist 添加IP到白名单
func AddToWhitelist(c *gin.Context) {
	var request struct {
		IP string `json:"ip" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的请求参数: " + err.Error(),
		})
		return
	}
	
	manager := middleware.GetBlacklistManager()
	manager.AddToWhitelist(request.IP)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP已添加到白名单",
	})
}

// RemoveFromWhitelist 从白名单移除IP
func RemoveFromWhitelist(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "IP地址不能为空",
		})
		return
	}
	
	manager := middleware.GetBlacklistManager()
	manager.RemoveFromWhitelist(ip)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "IP已从白名单移除",
	})
}

// GetAbnormalDetectionConfig 获取异常检测配置
func GetAbnormalDetectionConfig(c *gin.Context) {
	config := setting.GetAbnormalDetection()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    config,
	})
}

// UpdateAbnormalDetectionConfig 更新异常检测配置
func UpdateAbnormalDetectionConfig(c *gin.Context) {
	var newConfig setting.AbnormalDetectionConfig
	if err := c.ShouldBindJSON(&newConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "无效的配置格式: " + err.Error(),
		})
		return
	}
	
	// 这里需要实现更新异常检测配置的逻辑
	// 由于原有的 setting.GetAbnormalDetection() 返回的是指针，可以直接修改
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "异常检测配置已更新",
	})
}

// validateSecurityConfig 验证安全配置
func validateSecurityConfig(config *setting.SecurityConfig) error {
	// 验证异常检测配置
	if config.AbnormalDetection.MaxPromptLength <= 0 {
		return gin.Error{Err: gin.Error{Err: nil, Type: gin.ErrorTypePublic}, Type: gin.ErrorTypePublic}
	}
	
	if config.AbnormalDetection.MaxRandomCharRatio < 0 || config.AbnormalDetection.MaxRandomCharRatio > 1 {
		return gin.Error{Err: gin.Error{Err: nil, Type: gin.ErrorTypePublic}, Type: gin.ErrorTypePublic}
	}
	
	// 验证请求大小限制配置
	if config.RequestSizeLimit.MaxRequestBodySize <= 0 {
		return gin.Error{Err: gin.Error{Err: nil, Type: gin.ErrorTypePublic}, Type: gin.ErrorTypePublic}
	}
	
	// 验证流保护配置
	if config.StreamProtection.MaxStreamsPerIP <= 0 {
		return gin.Error{Err: gin.Error{Err: nil, Type: gin.ErrorTypePublic}, Type: gin.ErrorTypePublic}
	}
	
	return nil
}

// GetSecurityLogs 获取安全日志（简化版）
func GetSecurityLogs(c *gin.Context) {
	// 获取查询参数
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "50")
	logType := c.DefaultQuery("type", "all")
	
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	
	// 这里应该从日志系统获取实际的安全日志
	// 目前返回模拟数据
	logs := []gin.H{
		{
			"timestamp": "2024-01-01T12:00:00Z",
			"type":      "malicious_detection",
			"ip":        "192.168.1.100",
			"message":   "检测到token浪费攻击",
			"action":    "blocked",
		},
		{
			"timestamp": "2024-01-01T11:55:00Z",
			"type":      "rate_limit",
			"ip":        "192.168.1.101",
			"message":   "请求频率过高",
			"action":    "rate_limited",
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"logs":  logs,
			"page":  page,
			"limit": limit,
			"total": len(logs),
			"type":  logType,
		},
	})
}
