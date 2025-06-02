package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"
	"tea-api/common"
	"tea-api/dto"
)

// 请求大小限制配置
const (
	MaxRequestBodySize       = 10 * 1024 * 1024 // 10MB 最大请求体大小
	MaxPromptLengthLimit     = 100000           // 最大 Prompt 长度
	MaxMessagesCount         = 100              // 最大消息数量
	MaxSingleMessageSize     = 50000            // 单条消息最大大小
	MaxTokensLimit           = 100000           // 最大 tokens 限制
	MaxRandomContentRatio    = 0.9              // 最大随机内容比例
)

// RequestSizeLimit 请求大小限制中间件
func RequestSizeLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 只检查 POST 请求
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		// 检查 Content-Length
		if c.Request.ContentLength > MaxRequestBodySize {
			abortWithSizeError(c, fmt.Sprintf("请求体大小超过限制: %d bytes", c.Request.ContentLength))
			return
		}

		// 读取并验证请求体
		if c.Request.Body != nil {
			body, err := io.ReadAll(io.LimitReader(c.Request.Body, MaxRequestBodySize+1))
			if err != nil {
				abortWithSizeError(c, "读取请求体失败")
				return
			}

			// 检查实际大小
			if len(body) > MaxRequestBodySize {
				abortWithSizeError(c, "请求体大小超过限制")
				return
			}

			// 重新设置请求体
			c.Request.Body = io.NopCloser(bytes.NewReader(body))

			// 验证请求内容
			if err := validateRequestContent(body, c.ClientIP()); err != nil {
				abortWithSizeError(c, err.Error())
				return
			}
		}

		c.Next()
	}
}

// validateRequestContent 验证请求内容
func validateRequestContent(body []byte, clientIP string) error {
	// 尝试解析为 OpenAI 请求
	var request dto.GeneralOpenAIRequest
	if err := json.Unmarshal(body, &request); err != nil {
		// 如果不是 JSON 或不是 OpenAI 格式，跳过验证
		return nil
	}

	// 验证消息数量
	if len(request.Messages) > MaxMessagesCount {
		common.SysLog(fmt.Sprintf("too many messages from %s: %d", clientIP, len(request.Messages)))
		return fmt.Errorf("消息数量超过限制: %d", len(request.Messages))
	}

	// 验证 max_tokens
	if request.MaxTokens > MaxTokensLimit {
		common.SysLog(fmt.Sprintf("max_tokens too large from %s: %d", clientIP, request.MaxTokens))
		return fmt.Errorf("max_tokens 超过限制: %d", request.MaxTokens)
	}

	// 验证每条消息
	totalPromptLength := 0
	for i, message := range request.Messages {
		content := message.StringContent()
		messageSize := len(content)
		totalPromptLength += messageSize

		// 检查单条消息大小
		if messageSize > MaxSingleMessageSize {
			common.SysLog(fmt.Sprintf("message too large from %s: message[%d] size=%d", clientIP, i, messageSize))
			return fmt.Errorf("第 %d 条消息大小超过限制: %d 字符", i+1, messageSize)
		}

		// 检查消息内容质量
		if err := validateMessageContent(content, clientIP, i); err != nil {
			return err
		}
	}

	// 检查总 Prompt 长度
	if totalPromptLength > MaxPromptLengthLimit {
		common.SysLog(fmt.Sprintf("total prompt too large from %s: %d", clientIP, totalPromptLength))
		return fmt.Errorf("总 Prompt 长度超过限制: %d 字符", totalPromptLength)
	}

	return nil
}

// validateMessageContent 验证消息内容质量
func validateMessageContent(content, clientIP string, messageIndex int) error {
	if len(content) == 0 {
		return nil
	}

	// 检测随机内容
	if isRandomContent(content) {
		common.SysLog(fmt.Sprintf("random content detected from %s: message[%d]", clientIP, messageIndex))
		return fmt.Errorf("检测到第 %d 条消息包含大量随机内容", messageIndex+1)
	}

	// 检测重复字符
	if hasExcessiveRepetition(content) {
		common.SysLog(fmt.Sprintf("excessive repetition detected from %s: message[%d]", clientIP, messageIndex))
		return fmt.Errorf("检测到第 %d 条消息包含过多重复内容", messageIndex+1)
	}

	// 检测异常字符比例
	if hasAbnormalCharRatio(content) {
		common.SysLog(fmt.Sprintf("abnormal char ratio detected from %s: message[%d]", clientIP, messageIndex))
		return fmt.Errorf("检测到第 %d 条消息字符分布异常", messageIndex+1)
	}

	return nil
}

// isRandomContent 检测是否为随机内容
func isRandomContent(content string) bool {
	if len(content) < 1000 {
		return false // 短内容不检测
	}

	// 计算字符分布熵
	charCount := make(map[rune]int)
	totalChars := 0
	
	for _, char := range content {
		if unicode.IsLetter(char) || unicode.IsDigit(char) {
			charCount[char]++
			totalChars++
		}
	}

	if totalChars < 100 {
		return false
	}

	// 计算字符分布的均匀性
	expectedFreq := float64(totalChars) / float64(len(charCount))
	variance := 0.0
	
	for _, count := range charCount {
		diff := float64(count) - expectedFreq
		variance += diff * diff
	}
	variance /= float64(len(charCount))

	// 如果方差很小，说明字符分布很均匀，可能是随机内容
	uniformityThreshold := expectedFreq * expectedFreq * 0.1
	return variance < uniformityThreshold && len(charCount) > 20
}

// hasExcessiveRepetition 检测过度重复
func hasExcessiveRepetition(content string) bool {
	if len(content) < 100 {
		return false
	}

	// 检测连续重复的子串
	maxRepeat := 0
	for i := 1; i <= len(content)/2; i++ {
		pattern := content[:i]
		count := 1
		pos := i
		
		for pos+i <= len(content) {
			if content[pos:pos+i] == pattern {
				count++
				pos += i
			} else {
				break
			}
		}
		
		if count > maxRepeat {
			maxRepeat = count
		}
		
		// 如果重复次数过多，认为是异常
		if count > 10 && i > 10 {
			return true
		}
	}

	return false
}

// hasAbnormalCharRatio 检测异常字符比例
func hasAbnormalCharRatio(content string) bool {
	if len(content) < 100 {
		return false
	}

	letters := 0
	digits := 0
	spaces := 0
	others := 0
	total := 0

	for _, char := range content {
		total++
		switch {
		case unicode.IsLetter(char):
			letters++
		case unicode.IsDigit(char):
			digits++
		case unicode.IsSpace(char):
			spaces++
		default:
			others++
		}
	}

	if total == 0 {
		return false
	}

	// 检查各种字符比例是否异常
	letterRatio := float64(letters) / float64(total)
	digitRatio := float64(digits) / float64(total)
	spaceRatio := float64(spaces) / float64(total)
	otherRatio := float64(others) / float64(total)

	// 异常情况：
	// 1. 数字比例过高 (>50%)
	// 2. 特殊字符比例过高 (>30%)
	// 3. 空格比例过低 (<5%) 且内容很长
	// 4. 字母比例过低 (<20%) 且不是代码
	return digitRatio > 0.5 || 
		   otherRatio > 0.3 || 
		   (spaceRatio < 0.05 && len(content) > 1000) ||
		   (letterRatio < 0.2 && !looksLikeCode(content))
}

// looksLikeCode 检测是否像代码
func looksLikeCode(content string) bool {
	codeIndicators := []string{
		"function", "class", "import", "export", "const", "let", "var",
		"def", "if", "else", "for", "while", "return", "try", "catch",
		"#include", "public", "private", "static", "void", "int", "string",
		"{", "}", "[", "]", "(", ")", ";", "//", "/*", "*/",
	}

	lowerContent := strings.ToLower(content)
	indicators := 0
	
	for _, indicator := range codeIndicators {
		if strings.Contains(lowerContent, indicator) {
			indicators++
		}
	}

	return indicators >= 3
}

// abortWithSizeError 中止请求并返回大小错误
func abortWithSizeError(c *gin.Context, message string) {
	c.JSON(http.StatusRequestEntityTooLarge, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "request_too_large",
			"code":    "content_size_exceeded",
		},
	})
	c.Abort()
}
