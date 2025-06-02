# Tea API 安全防护系统

## 概述

针对类似 `1.txt` 中的恶意脚本攻击，Tea API 实现了多层安全防护机制，能够有效检测和阻止 token 浪费攻击、高频请求攻击、流式请求滥用等恶意行为。

## 攻击特征分析

### 1.txt 脚本的攻击模式：
- **高频请求**：每 0.05 秒发送一次请求
- **超长 Prompt**：生成 50,000-100,000 字符的随机内容
- **流式请求滥用**：故意慢速读取流式响应，占用连接资源
- **资源耗尽**：通过长时间占用连接来消耗服务器资源

## 防护机制

### 1. IP 黑名单系统 (`middleware/ip_blacklist.go`)

**功能特点：**
- 自动检测恶意IP并加入黑名单
- 支持临时封禁和永久封禁
- 白名单机制保护内网IP
- 自动清理过期条目

**配置参数：**
```go
MaxStreamsPerIP        = 3              // 每个IP最大并发流数量
MaxStreamsPerUser      = 5              // 每个用户最大并发流数量
TempBlockDuration      = 1 * time.Hour  // 临时封禁时长
PermanentBlockDuration = 24 * time.Hour // 永久封禁时长
MaxViolations          = 5              // 最大违规次数
```

### 2. 异常行为检测 (`middleware/abnormal_detection.go`)

**检测能力：**
- 高频请求检测
- 超长 Prompt 检测
- 随机内容检测
- 流式请求滥用检测
- 可疑行为评分系统

**关键阈值：**
```go
MaxPromptLength        = 50000  // 最大 Prompt 长度
MaxRandomCharRatio     = 0.8    // 最大随机字符比例
MinRequestInterval     = 100    // 最小请求间隔(毫秒)
SuspiciousScoreLimit   = 100    // 可疑分数限制
MaxConcurrentStreams   = 5      // 最大并发流请求数
```

### 3. 请求大小限制 (`middleware/request_size_limit.go`)

**限制内容：**
- 请求体大小限制
- 单条消息大小限制
- 消息数量限制
- 内容质量检测

**配置参数：**
```go
MaxRequestBodySize    = 10 * 1024 * 1024 // 10MB 最大请求体大小
MaxPromptLength       = 100000           // 最大 Prompt 长度
MaxMessagesCount      = 100              // 最大消息数量
MaxSingleMessageSize  = 50000            // 单条消息最大大小
MaxTokensLimit        = 100000           // 最大 tokens 限制
```

### 4. 流保护系统 (`middleware/stream_protection.go`)

**保护机制：**
- 并发流数量限制
- 流空闲超时检测
- 慢客户端检测
- 传输速率监控

**配置参数：**
```go
MaxStreamsPerIP        = 3              // 每个IP最大并发流数量
StreamIdleTimeout      = 30 * time.Second // 流空闲超时
StreamMaxDuration      = 10 * time.Minute // 流最大持续时间
MinBytesPerSecond      = 10             // 最小字节/秒传输速率
```

## 使用方法

### 1. 启用安全防护

安全中间件已在 `main.go` 中按优先级顺序自动加载：

```go
// 安全中间件 - 按优先级顺序添加
server.Use(middleware.IPBlacklist())           // IP黑名单检查（最高优先级）
server.Use(middleware.RequestSizeLimit())      // 请求大小限制
server.Use(middleware.AbnormalDetection())     // 异常行为检测
server.Use(middleware.StreamProtection())      // 流保护
```

### 2. 管理API接口

#### 获取安全统计信息
```bash
GET /api/security/stats
```

#### 获取安全配置
```bash
GET /api/security/config
```

#### 更新安全配置
```bash
PUT /api/security/config
Content-Type: application/json

{
  "abnormal_detection": {
    "enabled": true,
    "max_prompt_length": 50000,
    "suspicious_score_limit": 100
  },
  "request_size_limit": {
    "enabled": true,
    "max_request_body_size": 10485760
  }
}
```

#### IP黑名单管理
```bash
# 添加IP到黑名单
POST /api/security/blacklist
{
  "ip": "192.168.1.100",
  "reason": "恶意攻击",
  "temporary": true
}

# 移除IP从黑名单
DELETE /api/security/blacklist/192.168.1.100

# 添加IP到白名单
POST /api/security/whitelist
{
  "ip": "192.168.1.200"
}
```

### 3. 配置文件

安全配置存储在 `setting/security_config.go` 中，支持动态修改：

```go
// 获取安全配置
config := setting.GetSecurityConfig()

// 检查是否启用某个功能
if setting.IsAbnormalDetectionEnabled() {
    // 异常检测已启用
}
```

## 防护效果

### 针对 1.txt 脚本的防护：

1. **IP黑名单**：检测到恶意行为后自动封禁IP
2. **高频检测**：0.05秒间隔的请求会被立即识别
3. **内容检测**：50,000+字符的随机内容会被拦截
4. **流保护**：慢速读取的流连接会被强制断开
5. **资源限制**：超过限制的请求会被直接拒绝

### 日志记录：

系统会记录所有安全事件：
```
abnormal high frequency from 192.168.1.100: 20 req/s
detected token wasting attack from 192.168.1.100: prompt_len=75000, random_chars=true, stream=true
added IP 192.168.1.100 to blacklist: 检测到恶意行为：token浪费攻击 (temporary: true)
```

## 性能影响

- **IP黑名单检查**：O(1) 时间复杂度，几乎无性能影响
- **内容分析**：仅对大请求进行分析，正常请求无影响
- **流监控**：异步处理，不影响响应时间
- **内存使用**：合理的缓存策略，内存占用可控

## 自定义配置

可以根据实际需求调整各项阈值：

```go
// 修改异常检测阈值
const (
    MaxPromptLength      = 30000  // 降低最大Prompt长度
    SuspiciousScoreLimit = 50     // 降低可疑分数阈值
)

// 修改流保护参数
const (
    MaxStreamsPerIP   = 2         // 更严格的流数量限制
    StreamIdleTimeout = 15 * time.Second // 更短的空闲超时
)
```

## 监控和告警

建议配置监控系统关注以下指标：
- 被阻止的请求数量
- 黑名单IP数量
- 异常检测触发次数
- 活跃流连接数量

通过 `/api/security/stats` 接口可以获取实时统计数据。

## 注意事项

1. **白名单配置**：确保将合法的内网IP和管理IP加入白名单
2. **阈值调整**：根据业务特点调整各项阈值，避免误杀
3. **日志监控**：定期检查安全日志，及时发现新的攻击模式
4. **性能测试**：在生产环境部署前进行充分的性能测试

## 总结

通过多层防护机制，Tea API 能够有效防御类似 1.txt 脚本的恶意攻击，保护系统资源和用户数据安全。系统具有良好的可配置性和扩展性，可以根据实际需求进行调整和优化。
