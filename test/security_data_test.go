package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"tea-api/controller"
	"tea-api/middleware"
	"tea-api/setting"
)

func TestSecurityDataIntegrity(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// 添加安全路由
	securityGroup := router.Group("/api/security")
	{
		securityGroup.GET("/stats", controller.GetSecurityStats)
		securityGroup.GET("/blacklist", controller.GetBlacklist)
		securityGroup.GET("/logs", controller.GetSecurityLogs)
		securityGroup.POST("/blacklist", controller.AddToBlacklist)
	}

	// 测试1: 添加IP到黑名单并验证数据
	t.Run("TestBlacklistData", func(t *testing.T) {
		// 添加测试IP到黑名单
		manager := middleware.GetBlacklistManager()
		manager.AddToBlacklist("192.168.1.100", "测试封禁", true)
		
		// 请求黑名单数据
		req, _ := http.NewRequest("GET", "/api/security/blacklist", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}
		
		data := response["data"].(map[string]interface{})
		blacklist := data["blacklist"].([]interface{})
		
		// 验证黑名单数据
		if len(blacklist) == 0 {
			t.Error("Expected blacklist data, got empty array")
		}
		
		found := false
		for _, item := range blacklist {
			entry := item.(map[string]interface{})
			if entry["ip"] == "192.168.1.100" && entry["reason"] == "测试封禁" {
				found = true
				break
			}
		}
		
		if !found {
			t.Error("Test IP not found in blacklist data")
		}
		
		fmt.Printf("✅ 黑名单数据测试通过: 找到测试IP %s\n", "192.168.1.100")
	})

	// 测试2: 验证安全日志数据
	t.Run("TestSecurityLogs", func(t *testing.T) {
		// 添加测试日志
		setting.AddSecurityLog("test_type", "192.168.1.101", "测试日志消息", "blocked", 
			map[string]interface{}{
				"test_field": "test_value",
			})
		
		// 请求安全日志
		req, _ := http.NewRequest("GET", "/api/security/logs?page=1&limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}
		
		data := response["data"].(map[string]interface{})
		logs := data["logs"].([]interface{})
		
		// 验证日志数据
		if len(logs) == 0 {
			t.Error("Expected log data, got empty array")
		}
		
		found := false
		for _, item := range logs {
			entry := item.(map[string]interface{})
			if entry["ip"] == "192.168.1.101" && entry["message"] == "测试日志消息" {
				found = true
				break
			}
		}
		
		if !found {
			t.Error("Test log entry not found in logs data")
		}
		
		fmt.Printf("✅ 安全日志数据测试通过: 找到测试日志条目\n")
	})

	// 测试3: 验证安全统计数据
	t.Run("TestSecurityStats", func(t *testing.T) {
		// 请求安全统计
		req, _ := http.NewRequest("GET", "/api/security/stats", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}
		
		data := response["data"].(map[string]interface{})
		
		// 验证统计数据结构
		requiredFields := []string{"security_stats", "blacklist_stats", "stream_stats", "config"}
		for _, field := range requiredFields {
			if _, exists := data[field]; !exists {
				t.Errorf("Missing required field: %s", field)
			}
		}
		
		fmt.Printf("✅ 安全统计数据测试通过: 包含所有必需字段\n")
	})

	// 测试4: 测试添加IP到黑名单的API
	t.Run("TestAddToBlacklistAPI", func(t *testing.T) {
		requestBody := `{
			"ip": "192.168.1.102",
			"reason": "API测试封禁",
			"temporary": true
		}`
		
		req, _ := http.NewRequest("POST", "/api/security/blacklist", strings.NewReader(requestBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
		
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse response: %v", err)
		}
		
		if !response["success"].(bool) {
			t.Error("Expected success=true, got false")
		}
		
		// 验证IP确实被添加到黑名单
		time.Sleep(100 * time.Millisecond) // 等待处理完成
		
		req2, _ := http.NewRequest("GET", "/api/security/blacklist", nil)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)
		
		var response2 map[string]interface{}
		json.Unmarshal(w2.Body.Bytes(), &response2)
		
		data := response2["data"].(map[string]interface{})
		blacklist := data["blacklist"].([]interface{})
		
		found := false
		for _, item := range blacklist {
			entry := item.(map[string]interface{})
			if entry["ip"] == "192.168.1.102" && entry["reason"] == "API测试封禁" {
				found = true
				break
			}
		}
		
		if !found {
			t.Error("API added IP not found in blacklist")
		}
		
		fmt.Printf("✅ 添加IP到黑名单API测试通过: IP %s 已成功添加\n", "192.168.1.102")
	})
}

func TestDataConsistency(t *testing.T) {
	fmt.Println("\n🔍 数据一致性测试:")
	
	// 测试黑名单管理器数据一致性
	manager := middleware.GetBlacklistManager()
	
	// 添加测试数据
	testIPs := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
	for i, ip := range testIPs {
		manager.AddToBlacklist(ip, fmt.Sprintf("测试原因%d", i+1), i%2 == 0)
	}
	
	// 获取统计数据
	stats := middleware.GetBlacklistStats()
	blacklistData := manager.GetBlacklistData()
	
	// 验证数据一致性
	totalFromStats := stats["total_blacklisted"].(int)
	totalFromData := len(blacklistData)
	
	if totalFromStats != totalFromData {
		t.Errorf("数据不一致: 统计显示 %d 个IP，实际数据有 %d 个IP", totalFromStats, totalFromData)
	} else {
		fmt.Printf("✅ 黑名单数据一致性验证通过: %d 个IP\n", totalFromStats)
	}
	
	// 验证临时/永久封禁统计
	tempCount := stats["temporary_blocks"].(int)
	permCount := stats["permanent_blocks"].(int)
	
	actualTempCount := 0
	actualPermCount := 0
	for _, data := range blacklistData {
		if data["is_temporary"].(bool) {
			actualTempCount++
		} else {
			actualPermCount++
		}
	}
	
	if tempCount != actualTempCount || permCount != actualPermCount {
		t.Errorf("封禁类型统计不一致: 统计(临时:%d,永久:%d) vs 实际(临时:%d,永久:%d)", 
			tempCount, permCount, actualTempCount, actualPermCount)
	} else {
		fmt.Printf("✅ 封禁类型统计一致性验证通过: 临时 %d 个，永久 %d 个\n", tempCount, permCount)
	}
}

func TestLogDataIntegrity(t *testing.T) {
	fmt.Println("\n📝 日志数据完整性测试:")
	
	// 添加不同类型的测试日志
	testLogs := []struct {
		logType string
		ip      string
		message string
		action  string
	}{
		{"malicious_detection", "192.168.1.200", "检测到恶意行为", "blocked"},
		{"rate_limit", "192.168.1.201", "请求频率过高", "rate_limited"},
		{"ip_blacklist", "192.168.1.202", "IP已加入黑名单", "blacklisted"},
	}
	
	for _, log := range testLogs {
		setting.AddSecurityLog(log.logType, log.ip, log.message, log.action, 
			map[string]interface{}{
				"test": true,
				"timestamp": time.Now().Unix(),
			})
	}
	
	// 验证日志数据
	logs, total := setting.GetSecurityLogs(1, 10, "all", "")
	
	if total < len(testLogs) {
		t.Errorf("日志数量不足: 期望至少 %d 条，实际 %d 条", len(testLogs), total)
	}
	
	// 验证最近添加的日志
	foundCount := 0
	for _, log := range logs {
		for _, testLog := range testLogs {
			if log.IP == testLog.ip && log.Message == testLog.message {
				foundCount++
				break
			}
		}
	}
	
	if foundCount != len(testLogs) {
		t.Errorf("测试日志未完全找到: 期望 %d 条，找到 %d 条", len(testLogs), foundCount)
	} else {
		fmt.Printf("✅ 日志数据完整性验证通过: %d 条测试日志全部找到\n", foundCount)
	}
	
	// 测试日志过滤功能
	filteredLogs, _ := setting.GetSecurityLogs(1, 10, "malicious_detection", "")
	maliciousCount := 0
	for _, log := range filteredLogs {
		if log.Type == "malicious_detection" {
			maliciousCount++
		}
	}
	
	if maliciousCount == 0 {
		t.Error("日志类型过滤功能异常: 未找到恶意检测日志")
	} else {
		fmt.Printf("✅ 日志过滤功能验证通过: 找到 %d 条恶意检测日志\n", maliciousCount)
	}
}
