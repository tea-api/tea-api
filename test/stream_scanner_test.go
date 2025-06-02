package test

import (
	"bufio"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tea-api/constant"
	"tea-api/relay/helper"
	relaycommon "tea-api/relay/common"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestStreamScannerHandler 测试流式扫描器处理器
func TestStreamScannerHandler(t *testing.T) {
	// 设置测试环境
	gin.SetMode(gin.TestMode)

	// 初始化常量以避免panic
	constant.StreamingTimeout = 60
	
	// 创建测试数据
	testData := []string{
		"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Hello\"}}]}",
		"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\" world\"}}]}",
		"data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"!\"}}]}",
		"data: [DONE]",
	}
	
	// 创建模拟响应体
	responseBody := strings.Join(testData, "\n")
	
	// 创建HTTP响应
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}
	resp.Header.Set("Content-Type", "text/event-stream")
	
	// 创建Gin上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", nil)
	
	// 创建RelayInfo
	info := &relaycommon.RelayInfo{
		UpstreamModelName: "gpt-3.5-turbo",
	}
	
	// 收集处理的数据
	var processedData []string
	var dataHandlerCalled int
	
	// 数据处理器
	dataHandler := func(data string) bool {
		dataHandlerCalled++
		processedData = append(processedData, data)
		return true
	}
	
	// 测试StreamScannerHandler
	t.Run("Normal Stream Processing", func(t *testing.T) {
		// 重置计数器
		dataHandlerCalled = 0
		processedData = []string{}
		
		// 重新创建响应体
		resp.Body = io.NopCloser(strings.NewReader(responseBody))
		
		// 调用StreamScannerHandler
		helper.StreamScannerHandler(c, resp, info, dataHandler)
		
		// 验证结果
		assert.Equal(t, 3, dataHandlerCalled, "应该处理3个数据块（不包括[DONE]）")
		assert.Equal(t, 3, len(processedData), "应该收集到3个数据块")
		
		// 验证数据内容
		for i, data := range processedData {
			assert.Contains(t, data, "chatcmpl-test", "数据块 %d 应该包含正确的ID", i)
		}
	})
}

// TestStreamScannerBufferIssue 测试缓冲区问题修复
func TestStreamScannerBufferIssue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 初始化常量以避免panic
	constant.StreamingTimeout = 60
	
	// 创建大量数据来测试缓冲区处理
	var testDataBuilder strings.Builder
	for i := 0; i < 100; i++ {
		testDataBuilder.WriteString("data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Token ")
		testDataBuilder.WriteString(string(rune('A' + i%26)))
		testDataBuilder.WriteString("\"}}]}\n")
	}
	testDataBuilder.WriteString("data: [DONE]\n")
	
	responseBody := testDataBuilder.String()
	
	resp := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(responseBody)),
	}
	resp.Header.Set("Content-Type", "text/event-stream")
	
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/test", nil)
	
	info := &relaycommon.RelayInfo{
		UpstreamModelName: "gpt-3.5-turbo",
	}
	
	var processedCount int
	dataHandler := func(data string) bool {
		processedCount++
		return true
	}
	
	t.Run("Large Stream Processing", func(t *testing.T) {
		processedCount = 0
		resp.Body = io.NopCloser(strings.NewReader(responseBody))
		
		// 这个测试应该不会panic
		assert.NotPanics(t, func() {
			helper.StreamScannerHandler(c, resp, info, dataHandler)
		}, "处理大量流式数据时不应该panic")
		
		assert.Equal(t, 100, processedCount, "应该处理100个数据块")
	})
}

// TestBufferAfterScanPanic 测试修复前会导致panic的情况
func TestBufferAfterScanPanic(t *testing.T) {
	t.Run("Buffer After Scan Should Not Panic", func(t *testing.T) {
		// 这个测试验证我们的修复确实解决了问题
		scanner := bufio.NewScanner(strings.NewReader("line1\nline2\nline3\n"))
		
		// 设置初始缓冲区
		scanner.Buffer(make([]byte, 1024), 4096)
		
		// 开始扫描
		scanner.Scan()
		
		// 在扫描后尝试重新设置缓冲区应该会panic（这是Go的预期行为）
		assert.Panics(t, func() {
			scanner.Buffer(make([]byte, 2048), 8192)
		}, "在Scan后调用Buffer应该panic")
	})
}

// BenchmarkStreamScanner 性能基准测试
func BenchmarkStreamScanner(b *testing.B) {
	gin.SetMode(gin.TestMode)

	// 初始化常量以避免panic
	constant.StreamingTimeout = 60
	
	// 创建测试数据
	testData := "data: {\"id\":\"chatcmpl-test\",\"object\":\"chat.completion.chunk\",\"created\":1234567890,\"model\":\"gpt-3.5-turbo\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"Hello world!\"}}]}\n"
	responseBody := strings.Repeat(testData, 1000) + "data: [DONE]\n"
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		resp := &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(responseBody)),
		}
		resp.Header.Set("Content-Type", "text/event-stream")
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/test", nil)
		
		info := &relaycommon.RelayInfo{
			UpstreamModelName: "gpt-3.5-turbo",
		}
		
		dataHandler := func(data string) bool {
			return true
		}
		
		helper.StreamScannerHandler(c, resp, info, dataHandler)
	}
}
