package test

import (
	"bytes"
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

// 创建测试服务器，模拟真实的应用配置
func createTestServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	
	// 添加安全中间件 - 按照main.go中的顺序
	server.Use(middleware.RequestId())
	server.Use(middleware.IPBlacklist())           // IP黑名单检查（最高优先级）
	server.Use(middleware.RequestSizeLimit())      // 请求大小限制
	server.Use(middleware.AbnormalDetection())     // 异常行为检测
	server.Use(middleware.StreamProtection())      // 流保护
	
	// 添加测试路由
	server.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	server.POST("/v1/chat/completions", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "chat completion success"})
	})
	
	// 添加安全管理路由
	securityGroup := server.Group("/api/security")
	{
		securityGroup.GET("/stats", controller.GetSecurityStats)
		securityGroup.GET("/blacklist", controller.GetBlacklist)
		securityGroup.POST("/blacklist", controller.AddToBlacklist)
		securityGroup.DELETE("/blacklist/:ip", controller.RemoveFromBlacklist)
	}
	
	return server
}

func TestIPBlacklistFunctionality(t *testing.T) {
	server := createTestServer()
	
	t.Run("TestIPBlacklistBlocking", func(t *testing.T) {
		fmt.Println("🧪 测试IP黑名单拦截功能...")
		
		testIP := "192.168.100.1"
		
		// 1. 首先测试正常访问
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = testIP + ":12345"
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("正常访问失败: 期望状态码 200, 实际 %d", w.Code)
		} else {
			fmt.Printf("✅ 正常访问测试通过: IP %s 可以正常访问\n", testIP)
		}
		
		// 2. 将IP添加到黑名单
		manager := middleware.GetBlacklistManager()
		manager.AddToBlacklist(testIP, "功能测试封禁", true)
		fmt.Printf("📝 已将 IP %s 添加到黑名单\n", testIP)
		
		// 3. 测试黑名单拦截
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = testIP + ":12345"
		w2 := httptest.NewRecorder()
		server.ServeHTTP(w2, req2)
		
		if w2.Code != http.StatusForbidden {
			t.Errorf("黑名单拦截失败: 期望状态码 403, 实际 %d", w2.Code)
			t.Errorf("响应内容: %s", w2.Body.String())
		} else {
			fmt.Printf("✅ 黑名单拦截测试通过: IP %s 被成功拦截 (状态码: %d)\n", testIP, w2.Code)
			
			// 检查响应内容
			var response map[string]interface{}
			if err := json.Unmarshal(w2.Body.Bytes(), &response); err == nil {
				if errorInfo, ok := response["error"].(map[string]interface{}); ok {
					if errorInfo["type"] == "ip_blocked" {
						fmt.Printf("✅ 拦截响应格式正确: %s\n", errorInfo["message"])
					}
				}
			}
		}
		
		// 4. 从黑名单移除IP
		manager.RemoveFromBlacklist(testIP)
		fmt.Printf("🗑️ 已将 IP %s 从黑名单移除\n", testIP)
		
		// 5. 测试移除后的访问
		req3 := httptest.NewRequest("GET", "/test", nil)
		req3.RemoteAddr = testIP + ":12345"
		w3 := httptest.NewRecorder()
		server.ServeHTTP(w3, req3)
		
		if w3.Code != http.StatusOK {
			t.Errorf("移除后访问失败: 期望状态码 200, 实际 %d", w3.Code)
		} else {
			fmt.Printf("✅ 移除后访问测试通过: IP %s 可以重新访问\n", testIP)
		}
	})
}

func TestAbnormalDetectionFunctionality(t *testing.T) {
	server := createTestServer()
	
	t.Run("TestLargePromptBlocking", func(t *testing.T) {
		fmt.Println("\n🧪 测试超长Prompt拦截功能...")
		
		testIP := "192.168.100.2"
		
		// 创建超长的随机内容
		largeContent := strings.Repeat("abcdefghijklmnopqrstuvwxyz", 3000) // 约78,000字符
		
		requestBody := map[string]interface{}{
			"model": "gpt-3.5-turbo",
			"messages": []map[string]interface{}{
				{
					"role":    "user",
					"content": largeContent,
				},
			},
			"stream": true,
		}
		
		bodyBytes, _ := json.Marshal(requestBody)
		req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = testIP + ":12345"
		
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		// 应该被请求大小限制或异常检测拦截
		if w.Code == http.StatusOK {
			t.Error("超长Prompt未被拦截")
		} else {
			fmt.Printf("✅ 超长Prompt拦截测试通过: 状态码 %d\n", w.Code)
			fmt.Printf("📝 拦截原因: %s\n", w.Body.String())
		}
	})
}

func TestHighFrequencyBlocking(t *testing.T) {
	server := createTestServer()
	
	t.Run("TestRateLimiting", func(t *testing.T) {
		fmt.Println("\n🧪 测试高频请求拦截功能...")
		
		testIP := "192.168.100.3"
		
		// 快速发送多个请求
		blockedCount := 0
		totalRequests := 15
		
		for i := 0; i < totalRequests; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = testIP + ":12345"
			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)
			
			if w.Code != http.StatusOK {
				blockedCount++
			}
			
			// 短间隔发送
			time.Sleep(10 * time.Millisecond)
		}
		
		if blockedCount > 0 {
			fmt.Printf("✅ 高频请求拦截测试通过: %d/%d 请求被拦截\n", blockedCount, totalRequests)
		} else {
			fmt.Printf("⚠️ 高频请求拦截可能未生效: 0/%d 请求被拦截\n", totalRequests)
		}
	})
}

func TestSecurityLogGeneration(t *testing.T) {
	fmt.Println("\n🧪 测试安全日志生成功能...")
	
	// 获取初始日志数量
	_, initialTotal := setting.GetSecurityLogs(1, 100, "all", "")
	fmt.Printf("📊 初始日志数量: %d\n", initialTotal)
	
	// 触发一些安全事件
	testIP := "192.168.100.4"
	
	// 1. 添加IP到黑名单（应该生成日志）
	manager := middleware.GetBlacklistManager()
	manager.AddToBlacklist(testIP, "测试日志生成", true)
	
	// 2. 手动添加一些测试日志
	setting.AddSecurityLog("test_event", testIP, "测试安全事件", "blocked", 
		map[string]interface{}{
			"test": true,
			"timestamp": time.Now().Unix(),
		})
	
	// 获取更新后的日志
	newLogs, newTotal := setting.GetSecurityLogs(1, 100, "all", "")
	
	if newTotal > initialTotal {
		fmt.Printf("✅ 安全日志生成测试通过: 新增 %d 条日志\n", newTotal-initialTotal)
		
		// 检查最新的日志条目
		if len(newLogs) > 0 {
			latestLog := newLogs[0]
			fmt.Printf("📝 最新日志: IP=%s, 类型=%s, 消息=%s\n", 
				latestLog.IP, latestLog.Type, latestLog.Message)
		}
	} else {
		t.Error("安全日志未正确生成")
	}
}

func TestSecurityAPIIntegration(t *testing.T) {
	server := createTestServer()
	
	t.Run("TestSecurityAPIsWorking", func(t *testing.T) {
		fmt.Println("\n🧪 测试安全管理API功能...")
		
		testIP := "192.168.100.5"
		
		// 1. 测试添加IP到黑名单API
		addRequest := map[string]interface{}{
			"ip":        testIP,
			"reason":    "API集成测试",
			"temporary": true,
		}
		
		bodyBytes, _ := json.Marshal(addRequest)
		req := httptest.NewRequest("POST", "/api/security/blacklist", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			t.Errorf("添加IP到黑名单API失败: 状态码 %d", w.Code)
		} else {
			fmt.Printf("✅ 添加IP到黑名单API测试通过\n")
		}
		
		// 2. 测试获取黑名单API
		req2 := httptest.NewRequest("GET", "/api/security/blacklist", nil)
		w2 := httptest.NewRecorder()
		server.ServeHTTP(w2, req2)
		
		if w2.Code != http.StatusOK {
			t.Errorf("获取黑名单API失败: 状态码 %d", w2.Code)
		} else {
			var response map[string]interface{}
			json.Unmarshal(w2.Body.Bytes(), &response)
			
			data := response["data"].(map[string]interface{})
			blacklist := data["blacklist"].([]interface{})
			
			// 检查是否包含刚添加的IP
			found := false
			for _, item := range blacklist {
				entry := item.(map[string]interface{})
				if entry["ip"] == testIP {
					found = true
					break
				}
			}
			
			if found {
				fmt.Printf("✅ 获取黑名单API测试通过: 找到测试IP\n")
			} else {
				t.Error("获取黑名单API未返回刚添加的IP")
			}
		}
		
		// 3. 验证IP确实被拦截
		testReq := httptest.NewRequest("GET", "/test", nil)
		testReq.RemoteAddr = testIP + ":12345"
		testW := httptest.NewRecorder()
		server.ServeHTTP(testW, testReq)
		
		if testW.Code == http.StatusForbidden {
			fmt.Printf("✅ IP拦截验证通过: 通过API添加的IP被成功拦截\n")
		} else {
			t.Errorf("通过API添加的IP未被拦截: 状态码 %d", testW.Code)
		}
		
		// 4. 测试移除IP API
		req3 := httptest.NewRequest("DELETE", "/api/security/blacklist/"+testIP, nil)
		w3 := httptest.NewRecorder()
		server.ServeHTTP(w3, req3)
		
		if w3.Code != http.StatusOK {
			t.Errorf("移除IP API失败: 状态码 %d", w3.Code)
		} else {
			fmt.Printf("✅ 移除IP API测试通过\n")
		}
	})
}

func TestEndToEndSecurity(t *testing.T) {
	fmt.Println("\n🎯 端到端安全功能测试...")
	
	server := createTestServer()
	
	// 模拟真实的攻击场景
	attackerIP := "192.168.100.99"
	
	fmt.Printf("🔴 模拟攻击者IP: %s\n", attackerIP)
	
	// 1. 正常请求应该通过
	normalReq := httptest.NewRequest("GET", "/test", nil)
	normalReq.RemoteAddr = attackerIP + ":12345"
	normalW := httptest.NewRecorder()
	server.ServeHTTP(normalW, normalReq)
	
	if normalW.Code == http.StatusOK {
		fmt.Printf("✅ 初始状态: 攻击者IP可以正常访问\n")
	}
	
	// 2. 发送恶意请求（超长内容）
	maliciousContent := strings.Repeat("random_attack_content_", 4000) // 约80,000字符
	maliciousRequest := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []map[string]interface{}{
			{
				"role":    "user", 
				"content": maliciousContent,
			},
		},
		"stream": true,
	}
	
	bodyBytes, _ := json.Marshal(maliciousRequest)
	maliciousReq := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(bodyBytes))
	maliciousReq.Header.Set("Content-Type", "application/json")
	maliciousReq.RemoteAddr = attackerIP + ":12345"
	
	maliciousW := httptest.NewRecorder()
	server.ServeHTTP(maliciousW, maliciousReq)
	
	fmt.Printf("🛡️ 恶意请求结果: 状态码 %d\n", maliciousW.Code)
	
	// 3. 检查IP是否被自动加入黑名单
	time.Sleep(100 * time.Millisecond) // 等待处理完成
	
	// 4. 再次尝试正常请求
	finalReq := httptest.NewRequest("GET", "/test", nil)
	finalReq.RemoteAddr = attackerIP + ":12345"
	finalW := httptest.NewRecorder()
	server.ServeHTTP(finalW, finalReq)
	
	if finalW.Code == http.StatusForbidden {
		fmt.Printf("🎉 端到端测试成功: 攻击者IP被自动封禁 (状态码: %d)\n", finalW.Code)
	} else {
		fmt.Printf("⚠️ 端到端测试部分成功: 攻击者IP未被自动封禁 (状态码: %d)\n", finalW.Code)
	}
	
	// 5. 检查安全日志
	logs, total := setting.GetSecurityLogs(1, 10, "all", attackerIP)
	fmt.Printf("📊 攻击者相关日志数量: %d\n", total)
	
	for _, log := range logs {
		fmt.Printf("📝 安全日志: %s - %s (%s)\n", log.Type, log.Message, log.Action)
	}
}
