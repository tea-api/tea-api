package common

import (
	"time"
)

// LatencyOptimizationConfig 首字时延优化配置
type LatencyOptimizationConfig struct {
	// HTTP客户端优化
	HTTPMaxIdleConns        int           // HTTP最大空闲连接数
	HTTPMaxIdleConnsPerHost int           // 每个主机最大空闲连接数
	HTTPMaxConnsPerHost     int           // 每个主机最大连接数
	HTTPIdleConnTimeout     time.Duration // HTTP空闲连接超时
	HTTPResponseHeaderTimeout time.Duration // HTTP响应头超时
	HTTPTLSHandshakeTimeout time.Duration // TLS握手超时
	
	// 流式响应优化
	StreamInitialBufferSize int           // 流式响应初始缓冲区大小
	StreamMaxBufferSize     int           // 流式响应最大缓冲区大小
	StreamFirstTokenBuffer  int           // 首字响应专用缓冲区大小
	StreamFlushInterval     time.Duration // 流式响应刷新间隔
	
	// 数据库连接池优化
	DBMaxIdleConns    int           // 数据库最大空闲连接数
	DBMaxOpenConns    int           // 数据库最大连接数
	DBConnMaxLifetime time.Duration // 数据库连接最大生命周期
	DBConnMaxIdleTime time.Duration // 数据库连接最大空闲时间
	
	// Redis连接池优化
	RedisPoolSize          int           // Redis连接池大小
	RedisMinIdleConns      int           // Redis最小空闲连接数
	RedisMaxConnAge        time.Duration // Redis连接最大生命周期
	RedisPoolTimeout       time.Duration // Redis连接池超时
	RedisIdleTimeout       time.Duration // Redis空闲连接超时
	RedisIdleCheckFreq     time.Duration // Redis空闲检查频率
	
	// 缓存优化
	EnableTokenCache       bool          // 启用Token缓存
	EnableUserCache        bool          // 启用用户缓存
	EnableChannelCache     bool          // 启用渠道缓存
	CacheExpiration        time.Duration // 缓存过期时间
	
	// 中间件优化
	SkipUnnecessaryChecks  bool          // 跳过不必要的检查
	OptimizeAuthFlow       bool          // 优化认证流程
	EnableFastPath         bool          // 启用快速路径
}

// DefaultLatencyOptimizationConfig 返回默认的首字时延优化配置
func DefaultLatencyOptimizationConfig() *LatencyOptimizationConfig {
	return &LatencyOptimizationConfig{
		// HTTP客户端优化
		HTTPMaxIdleConns:        100,
		HTTPMaxIdleConnsPerHost: 20,
		HTTPMaxConnsPerHost:     50,
		HTTPIdleConnTimeout:     90 * time.Second,
		HTTPResponseHeaderTimeout: 30 * time.Second,
		HTTPTLSHandshakeTimeout: 10 * time.Second,
		
		// 流式响应优化
		StreamInitialBufferSize: 4 * 1024,      // 4KB
		StreamMaxBufferSize:     1024 * 1024,   // 1MB
		StreamFirstTokenBuffer:  1 * 1024,      // 1KB
		StreamFlushInterval:     50 * time.Millisecond,
		
		// 数据库连接池优化
		DBMaxIdleConns:    50,
		DBMaxOpenConns:    200,
		DBConnMaxLifetime: 300 * time.Second,
		DBConnMaxIdleTime: 60 * time.Second,
		
		// Redis连接池优化
		RedisPoolSize:      20,
		RedisMinIdleConns:  5,
		RedisMaxConnAge:    300 * time.Second,
		RedisPoolTimeout:   4 * time.Second,
		RedisIdleTimeout:   300 * time.Second,
		RedisIdleCheckFreq: 60 * time.Second,
		
		// 缓存优化
		EnableTokenCache:   true,
		EnableUserCache:    true,
		EnableChannelCache: true,
		CacheExpiration:    300 * time.Second,
		
		// 中间件优化
		SkipUnnecessaryChecks: false,
		OptimizeAuthFlow:      true,
		EnableFastPath:        true,
	}
}

// GetLatencyOptimizationConfig 获取首字时延优化配置
func GetLatencyOptimizationConfig() *LatencyOptimizationConfig {
	config := DefaultLatencyOptimizationConfig()
	
	// 从环境变量读取配置
	config.HTTPMaxIdleConns = GetEnvOrDefault("HTTP_MAX_IDLE_CONNS", config.HTTPMaxIdleConns)
	config.HTTPMaxIdleConnsPerHost = GetEnvOrDefault("HTTP_MAX_IDLE_CONNS_PER_HOST", config.HTTPMaxIdleConnsPerHost)
	config.HTTPMaxConnsPerHost = GetEnvOrDefault("HTTP_MAX_CONNS_PER_HOST", config.HTTPMaxConnsPerHost)
	
	config.StreamInitialBufferSize = GetEnvOrDefault("STREAM_INITIAL_BUFFER_SIZE", config.StreamInitialBufferSize)
	config.StreamMaxBufferSize = GetEnvOrDefault("STREAM_MAX_BUFFER_SIZE", config.StreamMaxBufferSize)
	config.StreamFirstTokenBuffer = GetEnvOrDefault("STREAM_FIRST_TOKEN_BUFFER", config.StreamFirstTokenBuffer)
	
	config.DBMaxIdleConns = GetEnvOrDefault("DB_MAX_IDLE_CONNS", config.DBMaxIdleConns)
	config.DBMaxOpenConns = GetEnvOrDefault("DB_MAX_OPEN_CONNS", config.DBMaxOpenConns)
	
	config.RedisPoolSize = GetEnvOrDefault("REDIS_POOL_SIZE", config.RedisPoolSize)
	config.RedisMinIdleConns = GetEnvOrDefault("REDIS_MIN_IDLE_CONNS", config.RedisMinIdleConns)
	
	config.EnableTokenCache = GetEnvOrDefaultBool("ENABLE_TOKEN_CACHE", config.EnableTokenCache)
	config.EnableUserCache = GetEnvOrDefaultBool("ENABLE_USER_CACHE", config.EnableUserCache)
	config.EnableChannelCache = GetEnvOrDefaultBool("ENABLE_CHANNEL_CACHE", config.EnableChannelCache)
	
	config.OptimizeAuthFlow = GetEnvOrDefaultBool("OPTIMIZE_AUTH_FLOW", config.OptimizeAuthFlow)
	config.EnableFastPath = GetEnvOrDefaultBool("ENABLE_FAST_PATH", config.EnableFastPath)
	
	return config
}

// 全局配置实例
var latencyOptConfig *LatencyOptimizationConfig

// InitLatencyOptimization 初始化首字时延优化配置
func InitLatencyOptimization() {
	latencyOptConfig = GetLatencyOptimizationConfig()
	SysLog("首字时延优化配置已加载")
}

// GetLatencyOptConfig 获取全局首字时延优化配置
func GetLatencyOptConfig() *LatencyOptimizationConfig {
	if latencyOptConfig == nil {
		InitLatencyOptimization()
	}
	return latencyOptConfig
}
