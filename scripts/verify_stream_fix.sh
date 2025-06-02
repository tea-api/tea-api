#!/bin/bash

# Tea API 流式响应修复验证脚本
# 用于验证 "Buffer called after Scan" 问题是否已修复

set -e

echo "=== Tea API 流式响应修复验证 ==="
echo ""

# 检查是否在正确的目录
if [ ! -f "go.mod" ]; then
    echo "错误: 请在项目根目录运行此脚本"
    exit 1
fi

echo "1. 编译项目..."
if go build -o tea-api; then
    echo "✅ 编译成功"
else
    echo "❌ 编译失败"
    exit 1
fi

echo ""
echo "2. 运行流式扫描器测试..."
if go test ./test -v -run TestStreamScanner; then
    echo "✅ 流式扫描器测试通过"
else
    echo "❌ 流式扫描器测试失败"
    exit 1
fi

echo ""
echo "3. 检查修复的关键点..."

# 检查是否移除了动态缓冲区扩展
if grep -q "bufferExpanded" relay/helper/stream_scanner.go; then
    echo "❌ 仍然存在 bufferExpanded 变量"
    exit 1
else
    echo "✅ 已移除动态缓冲区扩展逻辑"
fi

# 检查是否只有一次Buffer调用
buffer_calls=$(grep -c "scanner.Buffer" relay/helper/stream_scanner.go || true)
if [ "$buffer_calls" -eq 1 ]; then
    echo "✅ Scanner.Buffer 只调用一次"
else
    echo "❌ Scanner.Buffer 调用次数异常: $buffer_calls"
    exit 1
fi

# 检查缓冲区大小配置
if grep -q "InitialScannerBufferSize = 8 << 10" relay/helper/stream_scanner.go; then
    echo "✅ 使用优化的缓冲区大小 (8KB)"
else
    echo "❌ 缓冲区大小配置不正确"
    exit 1
fi

echo ""
echo "4. 运行性能基准测试..."
if go test ./test -bench=BenchmarkStreamScanner -run=^$ -benchtime=1s; then
    echo "✅ 性能基准测试完成"
else
    echo "❌ 性能基准测试失败"
    exit 1
fi

echo ""
echo "=== 修复验证完成 ==="
echo ""
echo "✅ 所有检查都通过了！"
echo ""
echo "修复摘要:"
echo "- 移除了在扫描过程中重新设置缓冲区的逻辑"
echo "- 使用固定的8KB初始缓冲区大小，平衡首字响应和处理效率"
echo "- 保留了首字响应时间记录和立即刷新功能"
echo "- 修复了 'Buffer called after Scan' panic 问题"
echo ""
echo "现在可以安全地部署修复后的版本。"
