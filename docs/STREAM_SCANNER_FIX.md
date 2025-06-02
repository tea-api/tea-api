# Tea API 流式响应 "Buffer called after Scan" 问题修复

## 问题描述

在Tea API的流式响应处理中，出现了 `panic in gopool.RelayPool: Buffer called after Scan` 错误，导致：

1. 上游API正确响应（状态码200，Content-Type: text/event-stream）
2. 但流式响应处理过程中发生panic
3. 最终接收到0个数据块，造成空回复问题

## 问题根源

问题出现在 `relay/helper/stream_scanner.go` 文件中的动态缓冲区扩展逻辑：

```go
// 问题代码（已修复）
if !firstTokenSent {
    info.SetFirstResponseTime()
    firstTokenSent = true
    
    // 首字响应后扩展缓冲区以提高后续处理效率
    if !bufferExpanded {
        scanner.Buffer(make([]byte, InitialScannerBufferSize), MaxScannerBufferSize)
        bufferExpanded = true
    }
}
```

**根本原因：**
- `bufio.Scanner` 在开始扫描后（调用 `Scan()` 方法后），不能再调用 `Buffer()` 方法重新设置缓冲区
- 代码在第75行已经设置了一次缓冲区，然后在扫描过程中（第140行）又尝试重新设置缓冲区
- 这违反了 `bufio.Scanner` 的使用规则，导致 panic: "Buffer called after Scan"

## 修复方案

### 1. 移除动态缓冲区扩展

移除了在扫描过程中重新设置缓冲区的逻辑，改为使用固定的优化缓冲区大小：

```go
// 修复后的代码
if !firstTokenSent {
    info.SetFirstResponseTime()
    firstTokenSent = true
}
```

### 2. 优化缓冲区配置

调整了缓冲区大小配置，平衡首字响应和处理效率：

```go
const (
    // 优化缓冲区大小以降低首字时延
    InitialScannerBufferSize = 8 << 10   // 8KB (8*1024) - 平衡首字响应和处理效率
    MaxScannerBufferSize     = 1 << 20   // 1MB (1*1024*1024) - 最大缓冲区
    DefaultPingInterval      = 10 * time.Second

    // 流式响应优化相关常量
    StreamFlushInterval      = 50 * time.Millisecond // 流式响应刷新间隔
)
```

### 3. 保留关键优化功能

修复过程中保留了以下重要功能：
- 首字响应时间记录
- 立即刷新机制
- 流式响应优化

## 修复的文件

1. **relay/helper/stream_scanner.go**
   - 移除 `bufferExpanded` 变量
   - 移除动态缓冲区扩展逻辑
   - 优化缓冲区大小配置

2. **docs/LATENCY_OPTIMIZATION.md**
   - 更新缓冲区策略文档

3. **middleware/request_size_limit.go**
   - 修复常量名冲突（`MaxPromptLength` → `MaxPromptLengthLimit`）

4. **controller/security.go**
   - 修复未使用变量问题

## 验证结果

### 编译测试
```bash
✅ 编译成功
```

### 功能测试
```bash
✅ 流式扫描器测试通过
✅ 大量流式数据处理测试通过
```

### 代码检查
```bash
✅ 已移除动态缓冲区扩展逻辑
✅ Scanner.Buffer 只调用一次
✅ 使用优化的缓冲区大小 (8KB)
```

### 性能测试
```bash
✅ 性能基准测试完成
BenchmarkStreamScanner-10    13380    87761 ns/op
```

## 部署建议

1. **立即部署**：此修复解决了导致空回复的关键问题
2. **监控日志**：部署后监控是否还有 "Buffer called after Scan" 错误
3. **性能观察**：观察流式响应的首字时延和整体性能

## 预期效果

修复后应该能够：
1. ✅ 消除 "Buffer called after Scan" panic
2. ✅ 正常处理流式响应数据
3. ✅ 避免空回复问题
4. ✅ 保持良好的首字响应性能

## 技术细节

### bufio.Scanner 使用规则
- `Buffer()` 方法只能在开始扫描前调用一次
- 一旦调用了 `Scan()` 方法，就不能再调用 `Buffer()`
- 违反此规则会导致 panic

### 缓冲区大小选择
- **8KB 初始缓冲区**：平衡首字响应和处理效率
- **1MB 最大缓冲区**：处理大型流式响应
- 避免过小缓冲区导致频繁重新分配
- 避免过大缓冲区影响首字响应时间

## 相关文件

- `relay/helper/stream_scanner.go` - 主要修复文件
- `test/stream_scanner_test.go` - 测试文件
- `scripts/verify_stream_fix.sh` - 验证脚本
- `docs/LATENCY_OPTIMIZATION.md` - 优化文档

---

**修复完成时间：** 2025年6月2日  
**修复验证：** 所有测试通过  
**部署状态：** 准备就绪
