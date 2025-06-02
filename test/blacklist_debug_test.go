package test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"tea-api/middleware"
)

func TestBlacklistDebug(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	
	// 只添加IP黑名单中间件
	server.Use(middleware.IPBlacklist())
	
	server.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	
	testIP := "192.168.200.1"
	
	fmt.Printf("=== 测试IP黑名单功能 ===\n")
	
	// 1. 测试正常访问
	fmt.Printf("1. 测试正常访问...\n")
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = testIP + ":12345"
	w1 := httptest.NewRecorder()
	server.ServeHTTP(w1, req1)
	fmt.Printf("正常访问结果: 状态码 %d\n", w1.Code)
	
	// 2. 添加IP到黑名单
	fmt.Printf("\n2. 添加IP到黑名单...\n")
	manager := middleware.GetBlacklistManager()
	manager.AddToBlacklist(testIP, "调试测试", true)
	
	// 3. 检查黑名单状态
	fmt.Printf("\n3. 检查黑名单状态...\n")
	stats := middleware.GetBlacklistStats()
	fmt.Printf("黑名单统计: %+v\n", stats)
	
	blacklistData := manager.GetBlacklistData()
	fmt.Printf("黑名单数据: %+v\n", blacklistData)
	
	// 4. 测试黑名单拦截
	fmt.Printf("\n4. 测试黑名单拦截...\n")
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = testIP + ":12345"
	w2 := httptest.NewRecorder()
	server.ServeHTTP(w2, req2)
	fmt.Printf("黑名单拦截结果: 状态码 %d\n", w2.Code)
	fmt.Printf("响应内容: %s\n", w2.Body.String())
	
	if w2.Code == 403 {
		fmt.Printf("✅ IP黑名单功能正常工作\n")
	} else {
		fmt.Printf("❌ IP黑名单功能未正常工作\n")
	}
}
