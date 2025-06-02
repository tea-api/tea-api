package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"tea-api/common"
)

// StreamConnection 流连接跟踪
type StreamConnection struct {
	StartTime    time.Time
	LastActivity time.Time
	BytesSent    int64
	ClientIP     string
	UserID       int
	TokenID      int
	IsActive     bool
	Context      context.Context
	Cancel       context.CancelFunc
}

// StreamMonitor 流监控器
type StreamMonitor struct {
	mu          sync.RWMutex
	connections map[string]*StreamConnection
	maxStreams  map[string]int // 每个IP的最大流数量
}

var streamMonitor = &StreamMonitor{
	connections: make(map[string]*StreamConnection),
	maxStreams:  make(map[string]int),
}

// 流保护配置
const (
	MaxStreamsPerIP        = 3              // 每个IP最大并发流数量
	MaxStreamsPerUser      = 5              // 每个用户最大并发流数量
	StreamIdleTimeout      = 30 * time.Second // 流空闲超时
	StreamMaxDuration      = 10 * time.Minute // 流最大持续时间
	MinBytesPerSecond      = 10             // 最小字节/秒传输速率
	SlowClientTimeout      = 60 * time.Second // 慢客户端超时
)

// StreamProtection 流保护中间件
func StreamProtection() gin.HandlerFunc {
	// 启动清理协程
	go streamMonitor.cleanupRoutine()
	
	return func(c *gin.Context) {
		// 检查是否为流式请求
		if !isStreamRequest(c) {
			c.Next()
			return
		}

		clientIP := c.ClientIP()
		userID := c.GetInt("id")
		tokenID := c.GetInt("token_id")
		
		// 检查并发流限制
		if !streamMonitor.checkStreamLimits(clientIP, userID) {
			abortWithStreamError(c, "超过最大并发流数量限制")
			return
		}

		// 创建流连接跟踪
		connectionID := fmt.Sprintf("%s_%d_%d_%d", clientIP, userID, tokenID, time.Now().UnixNano())
		ctx, cancel := context.WithTimeout(c.Request.Context(), StreamMaxDuration)
		
		connection := &StreamConnection{
			StartTime:    time.Now(),
			LastActivity: time.Now(),
			ClientIP:     clientIP,
			UserID:       userID,
			TokenID:      tokenID,
			IsActive:     true,
			Context:      ctx,
			Cancel:       cancel,
		}

		// 注册连接
		streamMonitor.registerConnection(connectionID, connection)
		defer func() {
			connection.Cancel()
			streamMonitor.unregisterConnection(connectionID)
		}()

		// 设置上下文
		c.Request = c.Request.WithContext(ctx)
		c.Set("stream_connection_id", connectionID)
		c.Set("stream_monitor", streamMonitor)

		// 包装 ResponseWriter 来监控传输
		wrappedWriter := &streamResponseWriter{
			ResponseWriter: c.Writer,
			connection:     connection,
			connectionID:   connectionID,
			monitor:        streamMonitor,
		}
		c.Writer = wrappedWriter

		c.Next()
	}
}

// isStreamRequest 检查是否为流式请求
func isStreamRequest(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return accept == "text/event-stream" || 
		   c.GetHeader("Cache-Control") == "no-cache" ||
		   c.Query("stream") == "true"
}

// checkStreamLimits 检查流限制
func (sm *StreamMonitor) checkStreamLimits(clientIP string, userID int) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	ipCount := 0
	userCount := 0

	for _, conn := range sm.connections {
		if !conn.IsActive {
			continue
		}
		
		if conn.ClientIP == clientIP {
			ipCount++
		}
		if conn.UserID == userID && userID > 0 {
			userCount++
		}
	}

	if ipCount >= MaxStreamsPerIP {
		common.SysLog(fmt.Sprintf("IP %s exceeded max streams limit: %d", clientIP, ipCount))
		return false
	}

	if userID > 0 && userCount >= MaxStreamsPerUser {
		common.SysLog(fmt.Sprintf("User %d exceeded max streams limit: %d", userID, userCount))
		return false
	}

	return true
}

// registerConnection 注册连接
func (sm *StreamMonitor) registerConnection(id string, conn *StreamConnection) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.connections[id] = conn
	
	common.SysLog(fmt.Sprintf("registered stream connection %s from %s", id, conn.ClientIP))
}

// unregisterConnection 注销连接
func (sm *StreamMonitor) unregisterConnection(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if conn, exists := sm.connections[id]; exists {
		conn.IsActive = false
		delete(sm.connections, id)
		common.SysLog(fmt.Sprintf("unregistered stream connection %s", id))
	}
}

// updateActivity 更新活动时间
func (sm *StreamMonitor) updateActivity(id string, bytes int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if conn, exists := sm.connections[id]; exists {
		conn.LastActivity = time.Now()
		conn.BytesSent += bytes
	}
}

// cleanupRoutine 清理协程
func (sm *StreamMonitor) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanupStaleConnections()
	}
}

// cleanupStaleConnections 清理过期连接
func (sm *StreamMonitor) cleanupStaleConnections() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	now := time.Now()
	for id, conn := range sm.connections {
		// 检查空闲超时
		if now.Sub(conn.LastActivity) > StreamIdleTimeout {
			common.SysLog(fmt.Sprintf("closing idle stream connection %s", id))
			conn.Cancel()
			conn.IsActive = false
			delete(sm.connections, id)
			continue
		}

		// 检查最大持续时间
		if now.Sub(conn.StartTime) > StreamMaxDuration {
			common.SysLog(fmt.Sprintf("closing long-running stream connection %s", id))
			conn.Cancel()
			conn.IsActive = false
			delete(sm.connections, id)
			continue
		}

		// 检查传输速率
		duration := now.Sub(conn.StartTime).Seconds()
		if duration > 10 && conn.BytesSent > 0 {
			rate := float64(conn.BytesSent) / duration
			if rate < MinBytesPerSecond {
				common.SysLog(fmt.Sprintf("closing slow stream connection %s (rate: %.2f bytes/s)", id, rate))
				conn.Cancel()
				conn.IsActive = false
				delete(sm.connections, id)
			}
		}
	}
}

// streamResponseWriter 流响应写入器
type streamResponseWriter struct {
	gin.ResponseWriter
	connection   *StreamConnection
	connectionID string
	monitor      *StreamMonitor
}

func (w *streamResponseWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	if n > 0 {
		w.monitor.updateActivity(w.connectionID, int64(n))
	}
	return n, err
}

func (w *streamResponseWriter) WriteString(s string) (int, error) {
	n, err := w.ResponseWriter.WriteString(s)
	if n > 0 {
		w.monitor.updateActivity(w.connectionID, int64(n))
	}
	return n, err
}

// abortWithStreamError 中止流请求
func abortWithStreamError(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error": gin.H{
			"message": message,
			"type":    "stream_limit_exceeded",
			"code":    "too_many_streams",
		},
	})
	c.Abort()
}

// GetStreamStats 获取流统计信息
func GetStreamStats() map[string]interface{} {
	streamMonitor.mu.RLock()
	defer streamMonitor.mu.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(streamMonitor.connections),
		"active_connections": 0,
		"connections_by_ip": make(map[string]int),
	}

	for _, conn := range streamMonitor.connections {
		if conn.IsActive {
			stats["active_connections"] = stats["active_connections"].(int) + 1
			ipCount := stats["connections_by_ip"].(map[string]int)[conn.ClientIP]
			stats["connections_by_ip"].(map[string]int)[conn.ClientIP] = ipCount + 1
		}
	}

	return stats
}
