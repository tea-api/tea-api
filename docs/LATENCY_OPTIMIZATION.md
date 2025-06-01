# Tea API 首字时延优化指南

本文档介绍如何通过代码层面的优化来降低 Tea API 的首字时延（Time To First Token, TTFT）。

## 🎯 优化目标

- **降低首字时延**：减少从请求发送到收到第一个token的时间
- **提升响应性能**：优化整体API响应速度
- **增强用户体验**：减少用户等待时间

## 🔧 优化内容

### 1. HTTP客户端优化

#### 连接池配置
- **连接复用**：启用HTTP Keep-Alive，减少连接建立时间
- **连接池大小**：优化最大连接数和空闲连接数
- **超时设置**：优化各种超时参数，特别是响应头超时

```go
// 优化后的HTTP传输层配置
transport := &http.Transport{
    MaxIdleConns:        100,              // 最大空闲连接数
    MaxIdleConnsPerHost: 20,               // 每个主机最大空闲连接数
    MaxConnsPerHost:     50,               // 每个主机最大连接数
    ResponseHeaderTimeout: 30 * time.Second, // 响应头超时
    // ... 其他优化配置
}
```

### 2. 流式响应优化

#### 缓冲区策略
- **动态缓冲区**：首字响应使用小缓冲区，后续扩展
- **立即刷新**：首字响应后立即刷新输出缓冲区
- **减少延迟**：优化数据处理流程

```go
// 首字响应优化
if !firstTokenSent {
    info.SetFirstResponseTime()
    firstTokenSent = true
    
    // 首字响应后扩展缓冲区
    scanner.Buffer(make([]byte, InitialScannerBufferSize), MaxScannerBufferSize)
    
    // 立即刷新
    if flusher, ok := c.Writer.(http.Flusher); ok {
        flusher.Flush()
    }
}
```

### 3. 数据库连接池优化

#### 连接管理
- **连接池大小**：根据并发量调整连接数
- **连接生命周期**：优化连接复用策略
- **空闲超时**：快速释放不活跃连接

### 4. Redis连接池优化

#### 缓存性能
- **连接池配置**：增加连接池大小和最小空闲连接
- **超时设置**：优化连接获取和空闲超时
- **检查频率**：定期清理过期连接

### 5. 快速认证缓存

#### 认证优化
- **内存缓存**：缓存认证结果，减少数据库查询
- **快速路径**：为已认证用户提供快速通道
- **过期管理**：自动清理过期缓存

## 📊 延迟监控

### 监控指标
- **认证延迟**：用户认证所需时间
- **连接延迟**：建立上游连接时间
- **首字时延**：从请求开始到首字响应的时间
- **总延迟**：请求完整处理时间

### 监控API
```bash
# 获取延迟统计
GET /api/latency/stats

# 重置统计数据
POST /api/latency/reset

# 获取优化配置
GET /api/latency/config
```

## ⚙️ 配置说明

### 环境变量配置

将以下配置添加到您的 `.env` 文件中：

```bash
# HTTP客户端优化
HTTP_MAX_IDLE_CONNS=100
HTTP_MAX_IDLE_CONNS_PER_HOST=20
HTTP_RESPONSE_HEADER_TIMEOUT=30

# 流式响应优化
STREAM_INITIAL_BUFFER_SIZE=4096
STREAM_FIRST_TOKEN_BUFFER=1024

# 数据库优化
SQL_MAX_IDLE_CONNS=50
SQL_MAX_OPEN_CONNS=200

# Redis优化
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=5

# 快速认证
ENABLE_FAST_PATH=true
OPTIMIZE_AUTH_FLOW=true
```

完整配置请参考 `docs/latency_optimization.env`

## 🚀 使用方法

### 1. 应用配置
```bash
# 复制优化配置
cp docs/latency_optimization.env .env.optimization

# 合并到现有配置
cat .env.optimization >> .env

# 或者选择性添加需要的配置项
```

### 2. 重启服务
```bash
# 重启 Tea API 服务
systemctl restart tea-api
# 或
docker-compose restart tea-api
```

### 3. 验证效果
```bash
# 测试首字时延
./bin/time_test.sh your-domain.com your-api-key 10

# 查看延迟统计
curl -H "Authorization: Bearer your-api-key" \
     https://your-domain.com/api/latency/stats
```

## 📈 性能调优建议

### 高并发场景
```bash
HTTP_MAX_IDLE_CONNS=200
HTTP_MAX_IDLE_CONNS_PER_HOST=50
SQL_MAX_OPEN_CONNS=300
REDIS_POOL_SIZE=30
```

### 低延迟场景
```bash
HTTP_RESPONSE_HEADER_TIMEOUT=10
STREAM_FIRST_TOKEN_BUFFER=512
STREAM_FLUSH_INTERVAL=25
ENABLE_FAST_PATH=true
```

### 资源受限场景
```bash
HTTP_MAX_IDLE_CONNS=50
SQL_MAX_OPEN_CONNS=100
REDIS_POOL_SIZE=10
STREAM_MAX_BUFFER_SIZE=524288
```

## ⚠️ 注意事项

1. **配置调优**：根据服务器资源和业务需求调整参数
2. **监控观察**：部署后持续监控延迟指标
3. **渐进优化**：建议逐步调整配置，观察效果
4. **测试验证**：在测试环境验证后再应用到生产环境

## 🔍 故障排查

### 常见问题

1. **首字时延仍然很高**
   - 检查上游API响应速度
   - 验证网络连接质量
   - 确认配置是否生效

2. **连接池耗尽**
   - 增加连接池大小
   - 检查连接泄漏
   - 优化连接生命周期

3. **内存使用增加**
   - 减少缓冲区大小
   - 调整缓存过期时间
   - 监控内存使用情况

### 调试方法
```bash
# 启用调试模式
DEBUG=true

# 查看详细日志
tail -f logs/tea-api.log | grep "延迟监控"

# 监控系统资源
htop
netstat -an | grep :3000
```

## 📚 相关文档

- [Tea API 部署指南](./DEPLOYMENT.md)
- [性能测试脚本](../bin/time_test.sh)
- [配置文件模板](./latency_optimization.env)
