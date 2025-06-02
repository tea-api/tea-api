package test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIPParsing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	server := gin.New()
	
	server.GET("/debug", func(c *gin.Context) {
		fmt.Printf("RemoteAddr: %s\n", c.Request.RemoteAddr)
		fmt.Printf("ClientIP(): %s\n", c.ClientIP())
		fmt.Printf("Header X-Real-IP: %s\n", c.GetHeader("X-Real-IP"))
		fmt.Printf("Header X-Forwarded-For: %s\n", c.GetHeader("X-Forwarded-For"))
		c.JSON(200, gin.H{
			"remote_addr": c.Request.RemoteAddr,
			"client_ip": c.ClientIP(),
		})
	})
	
	testIP := "192.168.100.1"
	
	// 测试1: 使用RemoteAddr
	req1 := httptest.NewRequest("GET", "/debug", nil)
	req1.RemoteAddr = testIP + ":12345"
	w1 := httptest.NewRecorder()
	
	fmt.Println("=== 测试1: RemoteAddr ===")
	server.ServeHTTP(w1, req1)
	
	// 测试2: 使用X-Real-IP头
	req2 := httptest.NewRequest("GET", "/debug", nil)
	req2.Header.Set("X-Real-IP", testIP)
	w2 := httptest.NewRecorder()
	
	fmt.Println("\n=== 测试2: X-Real-IP ===")
	server.ServeHTTP(w2, req2)
	
	// 测试3: 使用X-Forwarded-For头
	req3 := httptest.NewRequest("GET", "/debug", nil)
	req3.Header.Set("X-Forwarded-For", testIP)
	w3 := httptest.NewRecorder()
	
	fmt.Println("\n=== 测试3: X-Forwarded-For ===")
	server.ServeHTTP(w3, req3)
}
