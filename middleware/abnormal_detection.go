package middleware

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"tea-api/common"
	"tea-api/dto"
	"tea-api/setting"
)

type requestTracker struct {
	mu              sync.Mutex
	times           []time.Time
	totalRequests   int64
	streamRequests  int64
	largePrompts    int64
	suspiciousScore int64
	lastResetTime   time.Time
}

type requestMetrics struct {
	promptLength    int
	isStream        bool
	hasRandomChars  bool
	requestInterval time.Duration
}

var reqMap sync.Map

// 恶意行为检测阈值
const (
	MaxPromptLength        = 50000  // 最大 Prompt 长度
	MaxRandomCharRatio     = 0.8    // 最大随机字符比例
	MinRequestInterval     = 100    // 最小请求间隔(毫秒)
	SuspiciousScoreLimit   = 100    // 可疑分数限制
	MaxConcurrentStreams   = 5      // 最大并发流请求数
	StreamTimeoutSeconds   = 300    // 流请求超时时间(秒)
)

func AbnormalDetection() gin.HandlerFunc {
	return func(c *gin.Context) {
		config := setting.GetAbnormalDetection()
		if !config.Enabled {
			c.Next()
			return
		}

		identifier := c.ClientIP()
		val, _ := reqMap.LoadOrStore(identifier, &requestTracker{
			lastResetTime: time.Now(),
		})
		tracker := val.(*requestTracker)
		now := time.Now()

		// 分析请求内容
		metrics := analyzeRequest(c)

		tracker.mu.Lock()
		defer tracker.mu.Unlock()

		// 重置每小时统计
		if now.Sub(tracker.lastResetTime) > time.Hour {
			tracker.totalRequests = 0
			tracker.streamRequests = 0
			tracker.largePrompts = 0
			tracker.suspiciousScore = 0
			tracker.lastResetTime = now
		}

		tracker.times = append(tracker.times, now)
		tracker.totalRequests++

		// 清理旧的时间记录 (1秒窗口)
		window := now.Add(-1 * time.Second)
		i := 0
		for ; i < len(tracker.times); i++ {
			if tracker.times[i].After(window) {
				break
			}
		}
		tracker.times = tracker.times[i:]
		count := len(tracker.times)

		// 计算可疑分数
		suspiciousScore := calculateSuspiciousScore(metrics, tracker, count)
		tracker.suspiciousScore += suspiciousScore

		// 高频请求检测
		if config.Rules.HighFrequency.Enabled &&
			count > config.Rules.HighFrequency.MaxRequestsPerSecond {
			common.SysLog(fmt.Sprintf("abnormal high frequency from %s: %d req/s", identifier, count))
			tracker.suspiciousScore += 20

			if config.Security.SleepSeconds > 0 {
				time.Sleep(time.Duration(config.Security.SleepSeconds) * time.Second)
			}
		}

		// 恶意行为检测
		if detectMaliciousBehavior(metrics, tracker, identifier) {
			// 自动加入黑名单
			AutoBlacklistIP(identifier, "检测到恶意行为：token浪费攻击")
			abortWithMessage(c, "检测到恶意行为，请求被拒绝")
			return
		}

		// 可疑分数过高
		if tracker.suspiciousScore > SuspiciousScoreLimit {
			common.SysLog(fmt.Sprintf("suspicious score too high from %s: %d", identifier, tracker.suspiciousScore))
			// 临时封禁高可疑分数的IP
			AutoBlacklistIP(identifier, fmt.Sprintf("可疑行为分数过高：%d", tracker.suspiciousScore))
			abortWithMessage(c, "可疑行为分数过高，请求被限制")
			return
		}

		c.Next()
	}
}

// analyzeRequest 分析请求内容
func analyzeRequest(c *gin.Context) requestMetrics {
	metrics := requestMetrics{}

	// 检查是否为流式请求
	if c.Request.Header.Get("Accept") == "text/event-stream" ||
		c.GetHeader("Accept") == "text/event-stream" {
		metrics.isStream = true
	}

	// 分析请求体内容
	if c.Request.Body != nil && c.Request.Method == "POST" {
		body, err := io.ReadAll(c.Request.Body)
		if err == nil {
			// 重新设置请求体供后续使用
			c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

			// 解析 JSON 请求
			var request dto.GeneralOpenAIRequest
			if json.Unmarshal(body, &request) == nil {
				// 检查 stream 参数
				if request.Stream {
					metrics.isStream = true
				}

				// 分析 messages 内容
				totalLength := 0
				randomCharCount := 0
				totalCharCount := 0

				for _, message := range request.Messages {
					content := message.StringContent()
					totalLength += len(content)

					// 检测随机字符
					for _, char := range content {
						totalCharCount++
						if isRandomChar(char) {
							randomCharCount++
						}
					}
				}

				metrics.promptLength = totalLength
				if totalCharCount > 0 {
					randomRatio := float64(randomCharCount) / float64(totalCharCount)
					metrics.hasRandomChars = randomRatio > MaxRandomCharRatio
				}
			}
		}
	}

	return metrics
}

// isRandomChar 判断是否为随机字符
func isRandomChar(char rune) bool {
	// 检测连续的随机字母数字组合
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')
}

// calculateSuspiciousScore 计算可疑分数
func calculateSuspiciousScore(metrics requestMetrics, tracker *requestTracker, requestCount int) int64 {
	score := int64(0)

	// 超长 Prompt 加分
	if metrics.promptLength > MaxPromptLength {
		score += 30
		tracker.largePrompts++
	} else if metrics.promptLength > 20000 {
		score += 15
	}

	// 随机字符内容加分
	if metrics.hasRandomChars {
		score += 25
	}

	// 流式请求加分
	if metrics.isStream {
		score += 5
		tracker.streamRequests++
	}

	// 高频请求加分
	if requestCount > 10 {
		score += int64(requestCount - 10) * 2
	}

	// 流式请求过多加分
	if tracker.streamRequests > MaxConcurrentStreams {
		score += 20
	}

	return score
}

// detectMaliciousBehavior 检测恶意行为
func detectMaliciousBehavior(metrics requestMetrics, tracker *requestTracker, identifier string) bool {
	// 检测典型的 token 浪费攻击模式
	if metrics.promptLength > MaxPromptLength &&
		metrics.hasRandomChars &&
		metrics.isStream {
		common.SysLog(fmt.Sprintf("detected token wasting attack from %s: prompt_len=%d, random_chars=%v, stream=%v",
			identifier, metrics.promptLength, metrics.hasRandomChars, metrics.isStream))
		return true
	}

	// 检测过多的大 Prompt 请求
	if tracker.largePrompts > 10 {
		common.SysLog(fmt.Sprintf("too many large prompts from %s: %d", identifier, tracker.largePrompts))
		return true
	}

	// 检测过多的流式请求
	if tracker.streamRequests > MaxConcurrentStreams * 2 {
		common.SysLog(fmt.Sprintf("too many stream requests from %s: %d", identifier, tracker.streamRequests))
		return true
	}

	return false
}

// abortWithMessage 中止请求并返回错误消息
func abortWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "rate_limit_exceeded",
			"code":    "malicious_behavior_detected",
		},
	})
	c.Abort()
}
