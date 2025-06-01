package common

import (
	"fmt"
	"sync"
	"time"
)

// LatencyMetrics 延迟指标
type LatencyMetrics struct {
	RequestStartTime    time.Time     // 请求开始时间
	AuthCompleteTime    time.Time     // 认证完成时间
	UpstreamConnTime    time.Time     // 上游连接时间
	FirstTokenTime      time.Time     // 首字响应时间
	RequestCompleteTime time.Time     // 请求完成时间
	
	// 计算得出的延迟
	AuthLatency         time.Duration // 认证延迟
	ConnLatency         time.Duration // 连接延迟
	FirstTokenLatency   time.Duration // 首字时延 (TTFT)
	TotalLatency        time.Duration // 总延迟
}

// LatencyMonitor 延迟监控器
type LatencyMonitor struct {
	mu      sync.RWMutex
	metrics map[string]*LatencyMetrics
	
	// 统计信息
	totalRequests     int64
	avgFirstTokenLatency time.Duration
	maxFirstTokenLatency time.Duration
	minFirstTokenLatency time.Duration
}

var globalLatencyMonitor = &LatencyMonitor{
	metrics: make(map[string]*LatencyMetrics),
	minFirstTokenLatency: time.Hour, // 初始化为一个大值
}

// StartLatencyTracking 开始延迟跟踪
func StartLatencyTracking(requestID string) {
	globalLatencyMonitor.mu.Lock()
	defer globalLatencyMonitor.mu.Unlock()
	
	globalLatencyMonitor.metrics[requestID] = &LatencyMetrics{
		RequestStartTime: time.Now(),
	}
}

// RecordAuthComplete 记录认证完成时间
func RecordAuthComplete(requestID string) {
	globalLatencyMonitor.mu.Lock()
	defer globalLatencyMonitor.mu.Unlock()
	
	if metrics, exists := globalLatencyMonitor.metrics[requestID]; exists {
		metrics.AuthCompleteTime = time.Now()
		metrics.AuthLatency = metrics.AuthCompleteTime.Sub(metrics.RequestStartTime)
	}
}

// RecordUpstreamConnect 记录上游连接时间
func RecordUpstreamConnect(requestID string) {
	globalLatencyMonitor.mu.Lock()
	defer globalLatencyMonitor.mu.Unlock()
	
	if metrics, exists := globalLatencyMonitor.metrics[requestID]; exists {
		metrics.UpstreamConnTime = time.Now()
		metrics.ConnLatency = metrics.UpstreamConnTime.Sub(metrics.AuthCompleteTime)
	}
}

// RecordFirstToken 记录首字响应时间
func RecordFirstToken(requestID string) {
	globalLatencyMonitor.mu.Lock()
	defer globalLatencyMonitor.mu.Unlock()
	
	if metrics, exists := globalLatencyMonitor.metrics[requestID]; exists {
		metrics.FirstTokenTime = time.Now()
		metrics.FirstTokenLatency = metrics.FirstTokenTime.Sub(metrics.RequestStartTime)
		
		// 更新统计信息
		globalLatencyMonitor.updateStats(metrics.FirstTokenLatency)
	}
}

// RecordRequestComplete 记录请求完成时间
func RecordRequestComplete(requestID string) *LatencyMetrics {
	globalLatencyMonitor.mu.Lock()
	defer globalLatencyMonitor.mu.Unlock()
	
	if metrics, exists := globalLatencyMonitor.metrics[requestID]; exists {
		metrics.RequestCompleteTime = time.Now()
		metrics.TotalLatency = metrics.RequestCompleteTime.Sub(metrics.RequestStartTime)
		
		// 清理已完成的请求指标
		result := *metrics
		delete(globalLatencyMonitor.metrics, requestID)
		
		return &result
	}
	
	return nil
}

// updateStats 更新统计信息
func (lm *LatencyMonitor) updateStats(firstTokenLatency time.Duration) {
	lm.totalRequests++
	
	// 更新平均值
	if lm.totalRequests == 1 {
		lm.avgFirstTokenLatency = firstTokenLatency
	} else {
		// 使用移动平均
		lm.avgFirstTokenLatency = time.Duration(
			(int64(lm.avgFirstTokenLatency)*int64(lm.totalRequests-1) + int64(firstTokenLatency)) / int64(lm.totalRequests),
		)
	}
	
	// 更新最大值
	if firstTokenLatency > lm.maxFirstTokenLatency {
		lm.maxFirstTokenLatency = firstTokenLatency
	}
	
	// 更新最小值
	if firstTokenLatency < lm.minFirstTokenLatency {
		lm.minFirstTokenLatency = firstTokenLatency
	}
}

// GetLatencyStats 获取延迟统计信息
func GetLatencyStats() map[string]interface{} {
	globalLatencyMonitor.mu.RLock()
	defer globalLatencyMonitor.mu.RUnlock()
	
	return map[string]interface{}{
		"total_requests":           globalLatencyMonitor.totalRequests,
		"avg_first_token_latency":  globalLatencyMonitor.avgFirstTokenLatency.Milliseconds(),
		"max_first_token_latency":  globalLatencyMonitor.maxFirstTokenLatency.Milliseconds(),
		"min_first_token_latency":  globalLatencyMonitor.minFirstTokenLatency.Milliseconds(),
		"active_requests":          len(globalLatencyMonitor.metrics),
	}
}

// LogLatencyMetrics 记录延迟指标到日志
func LogLatencyMetrics(requestID string, metrics *LatencyMetrics) {
	if !DebugEnabled {
		return
	}
	
	logMsg := fmt.Sprintf(
		"[延迟监控] RequestID: %s, 认证延迟: %dms, 连接延迟: %dms, 首字时延: %dms, 总延迟: %dms",
		requestID,
		metrics.AuthLatency.Milliseconds(),
		metrics.ConnLatency.Milliseconds(),
		metrics.FirstTokenLatency.Milliseconds(),
		metrics.TotalLatency.Milliseconds(),
	)
	
	SysLog(logMsg)
	
	// 如果首字时延超过阈值，记录警告
	if metrics.FirstTokenLatency > 2*time.Second {
		SysLog(fmt.Sprintf("[警告] 首字时延过高: %dms, RequestID: %s", 
			metrics.FirstTokenLatency.Milliseconds(), requestID))
	}
}

// ResetLatencyStats 重置延迟统计
func ResetLatencyStats() {
	globalLatencyMonitor.mu.Lock()
	defer globalLatencyMonitor.mu.Unlock()
	
	globalLatencyMonitor.totalRequests = 0
	globalLatencyMonitor.avgFirstTokenLatency = 0
	globalLatencyMonitor.maxFirstTokenLatency = 0
	globalLatencyMonitor.minFirstTokenLatency = time.Hour
}

// CleanupStaleMetrics 清理过期的指标
func CleanupStaleMetrics() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			globalLatencyMonitor.mu.Lock()
			now := time.Now()
			for requestID, metrics := range globalLatencyMonitor.metrics {
				// 清理超过10分钟的未完成请求
				if now.Sub(metrics.RequestStartTime) > 10*time.Minute {
					delete(globalLatencyMonitor.metrics, requestID)
				}
			}
			globalLatencyMonitor.mu.Unlock()
		}
	}
}

// init 初始化延迟监控
func init() {
	go CleanupStaleMetrics()
}
