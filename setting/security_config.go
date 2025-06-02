package setting

import (
	"sync"
	"time"
	"tea-api/setting/config"
)

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 异常检测配置
	AbnormalDetection AbnormalDetectionSettings `json:"abnormal_detection"`
	
	// 请求大小限制配置
	RequestSizeLimit RequestSizeLimitSettings `json:"request_size_limit"`
	
	// 流保护配置
	StreamProtection StreamProtectionSettings `json:"stream_protection"`
	
	// IP黑名单配置
	IPBlacklist IPBlacklistSettings `json:"ip_blacklist"`
	
	// 内容过滤配置
	ContentFilter ContentFilterSettings `json:"content_filter"`
}

// AbnormalDetectionSettings 异常检测设置
type AbnormalDetectionSettings struct {
	Enabled                bool  `json:"enabled"`
	MaxPromptLength        int   `json:"max_prompt_length"`
	MaxRandomCharRatio     float64 `json:"max_random_char_ratio"`
	MinRequestInterval     int   `json:"min_request_interval_ms"`
	SuspiciousScoreLimit   int64 `json:"suspicious_score_limit"`
	MaxConcurrentStreams   int   `json:"max_concurrent_streams"`
	StreamTimeoutSeconds   int   `json:"stream_timeout_seconds"`
	AutoBlacklistEnabled   bool  `json:"auto_blacklist_enabled"`
}

// RequestSizeLimitSettings 请求大小限制设置
type RequestSizeLimitSettings struct {
	Enabled              bool `json:"enabled"`
	MaxRequestBodySize   int  `json:"max_request_body_size"`
	MaxPromptLength      int  `json:"max_prompt_length"`
	MaxMessagesCount     int  `json:"max_messages_count"`
	MaxSingleMessageSize int  `json:"max_single_message_size"`
	MaxTokensLimit       int  `json:"max_tokens_limit"`
	ContentValidation    bool `json:"content_validation"`
}

// StreamProtectionSettings 流保护设置
type StreamProtectionSettings struct {
	Enabled               bool `json:"enabled"`
	MaxStreamsPerIP       int  `json:"max_streams_per_ip"`
	MaxStreamsPerUser     int  `json:"max_streams_per_user"`
	StreamIdleTimeoutSec  int  `json:"stream_idle_timeout_sec"`
	StreamMaxDurationSec  int  `json:"stream_max_duration_sec"`
	MinBytesPerSecond     int  `json:"min_bytes_per_second"`
	SlowClientTimeoutSec  int  `json:"slow_client_timeout_sec"`
}

// IPBlacklistSettings IP黑名单设置
type IPBlacklistSettings struct {
	Enabled                bool `json:"enabled"`
	TempBlockDurationHours int  `json:"temp_block_duration_hours"`
	PermBlockDurationHours int  `json:"perm_block_duration_hours"`
	MaxViolations          int  `json:"max_violations"`
	CleanupIntervalMinutes int  `json:"cleanup_interval_minutes"`
	AutoBlacklistEnabled   bool `json:"auto_blacklist_enabled"`
}

// ContentFilterSettings 内容过滤设置
type ContentFilterSettings struct {
	Enabled                bool    `json:"enabled"`
	RandomContentDetection bool    `json:"random_content_detection"`
	RepetitionDetection    bool    `json:"repetition_detection"`
	CharRatioDetection     bool    `json:"char_ratio_detection"`
	MinContentLength       int     `json:"min_content_length"`
	MaxRepetitionRatio     float64 `json:"max_repetition_ratio"`
}

// 默认安全配置
var defaultSecurityConfig = SecurityConfig{
	AbnormalDetection: AbnormalDetectionSettings{
		Enabled:                true,
		MaxPromptLength:        50000,
		MaxRandomCharRatio:     0.8,
		MinRequestInterval:     100,
		SuspiciousScoreLimit:   100,
		MaxConcurrentStreams:   5,
		StreamTimeoutSeconds:   300,
		AutoBlacklistEnabled:   true,
	},
	RequestSizeLimit: RequestSizeLimitSettings{
		Enabled:              true,
		MaxRequestBodySize:   10 * 1024 * 1024, // 10MB
		MaxPromptLength:      100000,
		MaxMessagesCount:     100,
		MaxSingleMessageSize: 50000,
		MaxTokensLimit:       100000,
		ContentValidation:    true,
	},
	StreamProtection: StreamProtectionSettings{
		Enabled:               true,
		MaxStreamsPerIP:       3,
		MaxStreamsPerUser:     5,
		StreamIdleTimeoutSec:  30,
		StreamMaxDurationSec:  600, // 10 minutes
		MinBytesPerSecond:     10,
		SlowClientTimeoutSec:  60,
	},
	IPBlacklist: IPBlacklistSettings{
		Enabled:                true,
		TempBlockDurationHours: 1,
		PermBlockDurationHours: 24,
		MaxViolations:          5,
		CleanupIntervalMinutes: 10,
		AutoBlacklistEnabled:   true,
	},
	ContentFilter: ContentFilterSettings{
		Enabled:                true,
		RandomContentDetection: true,
		RepetitionDetection:    true,
		CharRatioDetection:     true,
		MinContentLength:       1000,
		MaxRepetitionRatio:     0.3,
	},
}

func init() {
	config.GlobalConfig.Register("security", &defaultSecurityConfig)
}

// GetSecurityConfig 获取安全配置
func GetSecurityConfig() *SecurityConfig {
	return &defaultSecurityConfig
}

// UpdateSecurityConfig 更新安全配置
func UpdateSecurityConfig(newConfig SecurityConfig) {
	defaultSecurityConfig = newConfig
}

// IsAbnormalDetectionEnabled 检查异常检测是否启用
func IsAbnormalDetectionEnabled() bool {
	return defaultSecurityConfig.AbnormalDetection.Enabled
}

// IsRequestSizeLimitEnabled 检查请求大小限制是否启用
func IsRequestSizeLimitEnabled() bool {
	return defaultSecurityConfig.RequestSizeLimit.Enabled
}

// IsStreamProtectionEnabled 检查流保护是否启用
func IsStreamProtectionEnabled() bool {
	return defaultSecurityConfig.StreamProtection.Enabled
}

// IsIPBlacklistEnabled 检查IP黑名单是否启用
func IsIPBlacklistEnabled() bool {
	return defaultSecurityConfig.IPBlacklist.Enabled
}

// IsContentFilterEnabled 检查内容过滤是否启用
func IsContentFilterEnabled() bool {
	return defaultSecurityConfig.ContentFilter.Enabled
}

// GetAbnormalDetectionSettings 获取异常检测设置
func GetAbnormalDetectionSettings() AbnormalDetectionSettings {
	return defaultSecurityConfig.AbnormalDetection
}

// GetRequestSizeLimitSettings 获取请求大小限制设置
func GetRequestSizeLimitSettings() RequestSizeLimitSettings {
	return defaultSecurityConfig.RequestSizeLimit
}

// GetStreamProtectionSettings 获取流保护设置
func GetStreamProtectionSettings() StreamProtectionSettings {
	return defaultSecurityConfig.StreamProtection
}

// GetIPBlacklistSettings 获取IP黑名单设置
func GetIPBlacklistSettings() IPBlacklistSettings {
	return defaultSecurityConfig.IPBlacklist
}

// GetContentFilterSettings 获取内容过滤设置
func GetContentFilterSettings() ContentFilterSettings {
	return defaultSecurityConfig.ContentFilter
}

// SecurityStats 安全统计信息
type SecurityStats struct {
	BlockedRequests      int64            `json:"blocked_requests"`
	MaliciousDetections  int64            `json:"malicious_detections"`
	BlacklistedIPs       int              `json:"blacklisted_ips"`
	ActiveStreams        int              `json:"active_streams"`
	SuspiciousActivities map[string]int64 `json:"suspicious_activities"`
}

var securityStats = &SecurityStats{
	SuspiciousActivities: make(map[string]int64),
}

// GetSecurityStats 获取安全统计信息
func GetSecurityStats() *SecurityStats {
	return securityStats
}

// IncrementBlockedRequests 增加被阻止的请求数
func IncrementBlockedRequests() {
	securityStats.BlockedRequests++
}

// IncrementMaliciousDetections 增加恶意检测数
func IncrementMaliciousDetections() {
	securityStats.MaliciousDetections++
}

// IncrementSuspiciousActivity 增加可疑活动计数
func IncrementSuspiciousActivity(activityType string) {
	securityStats.SuspiciousActivities[activityType]++
}

// UpdateBlacklistedIPs 更新黑名单IP数量
func UpdateBlacklistedIPs(count int) {
	securityStats.BlacklistedIPs = count
}

// UpdateActiveStreams 更新活跃流数量
func UpdateActiveStreams(count int) {
	securityStats.ActiveStreams = count
}

// SecurityLogEntry 安全日志条目
type SecurityLogEntry struct {
	ID        int64                  `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	IP        string                 `json:"ip"`
	Message   string                 `json:"message"`
	Action    string                 `json:"action"`
	Details   map[string]interface{} `json:"details"`
}

var securityLogs []SecurityLogEntry
var logMutex sync.RWMutex
var logIDCounter int64

// AddSecurityLog 添加安全日志
func AddSecurityLog(logType, ip, message, action string, details map[string]interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	logIDCounter++
	entry := SecurityLogEntry{
		ID:        logIDCounter,
		Timestamp: time.Now(),
		Type:      logType,
		IP:        ip,
		Message:   message,
		Action:    action,
		Details:   details,
	}

	securityLogs = append(securityLogs, entry)

	// 保持最多1000条日志
	if len(securityLogs) > 1000 {
		securityLogs = securityLogs[len(securityLogs)-1000:]
	}
}

// GetSecurityLogs 获取安全日志
func GetSecurityLogs(page, limit int, logType, ip string) ([]SecurityLogEntry, int) {
	logMutex.RLock()
	defer logMutex.RUnlock()

	// 过滤日志
	var filteredLogs []SecurityLogEntry
	for i := len(securityLogs) - 1; i >= 0; i-- {
		log := securityLogs[i]

		// 类型过滤
		if logType != "all" && log.Type != logType {
			continue
		}

		// IP过滤
		if ip != "" && log.IP != ip {
			continue
		}

		filteredLogs = append(filteredLogs, log)
	}

	total := len(filteredLogs)

	// 分页
	start := (page - 1) * limit
	end := start + limit

	if start >= total {
		return []SecurityLogEntry{}, total
	}

	if end > total {
		end = total
	}

	return filteredLogs[start:end], total
}
