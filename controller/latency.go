package controller

import (
	"net/http"
	"tea-api/common"

	"github.com/gin-gonic/gin"
)

// GetLatencyStats 获取延迟统计信息
func GetLatencyStats(c *gin.Context) {
	stats := common.GetLatencyStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    stats,
	})
}

// ResetLatencyStats 重置延迟统计
func ResetLatencyStats(c *gin.Context) {
	common.ResetLatencyStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "延迟统计已重置",
		"data":    nil,
	})
}

// GetLatencyOptimizationConfig 获取首字时延优化配置
func GetLatencyOptimizationConfig(c *gin.Context) {
	config := common.GetLatencyOptConfig()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    config,
	})
}
