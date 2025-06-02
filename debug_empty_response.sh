#!/bin/bash

# Tea API 空回复调试脚本
# 用于在服务器上诊断和解决空回复问题

echo "=== Tea API 空回复问题调试脚本 ==="
echo "此脚本将帮助您诊断和解决空回复问题"
echo ""

# 检查是否为root用户
if [ "$EUID" -ne 0 ]; then
    echo "请使用 sudo 运行此脚本"
    exit 1
fi

# 获取Tea API进程信息
echo "1. 检查Tea API服务状态..."
systemctl status tea-api || echo "systemctl 不可用，尝试其他方法..."

# 查找Tea API进程
TEA_PID=$(pgrep -f "tea-api" | head -1)
if [ -z "$TEA_PID" ]; then
    echo "❌ 未找到Tea API进程"
    echo "请确保Tea API正在运行"
    exit 1
else
    echo "✅ 找到Tea API进程: PID $TEA_PID"
fi

# 查找配置文件
echo ""
echo "2. 查找配置文件..."
CONFIG_PATHS=(
    "/opt/tea-api/config.json"
    "/etc/tea-api/config.json"
    "/usr/local/tea-api/config.json"
    "./config.json"
    "/root/tea-api/config.json"
)

CONFIG_FILE=""
for path in "${CONFIG_PATHS[@]}"; do
    if [ -f "$path" ]; then
        CONFIG_FILE="$path"
        echo "✅ 找到配置文件: $CONFIG_FILE"
        break
    fi
done

if [ -z "$CONFIG_FILE" ]; then
    echo "❌ 未找到配置文件，请手动指定配置文件路径"
    read -p "请输入配置文件路径: " CONFIG_FILE
    if [ ! -f "$CONFIG_FILE" ]; then
        echo "❌ 配置文件不存在: $CONFIG_FILE"
        exit 1
    fi
fi

# 备份原配置文件
echo ""
echo "3. 备份配置文件..."
cp "$CONFIG_FILE" "${CONFIG_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
echo "✅ 配置文件已备份"

# 启用调试模式
echo ""
echo "4. 启用调试模式..."
if grep -q '"debug"' "$CONFIG_FILE"; then
    # 如果已存在debug配置，修改为true
    sed -i 's/"debug"[[:space:]]*:[[:space:]]*false/"debug": true/g' "$CONFIG_FILE"
    sed -i 's/"debug"[[:space:]]*:[[:space:]]*"false"/"debug": "true"/g' "$CONFIG_FILE"
else
    # 如果不存在debug配置，添加到配置文件中
    sed -i '2i\  "debug": true,' "$CONFIG_FILE"
fi

# 设置日志目录
LOG_DIR="/var/log/tea-api"
if grep -q '"log_dir"' "$CONFIG_FILE"; then
    sed -i "s|\"log_dir\"[[:space:]]*:[[:space:]]*\"[^\"]*\"|\"log_dir\": \"$LOG_DIR\"|g" "$CONFIG_FILE"
else
    sed -i "3i\  \"log_dir\": \"$LOG_DIR\"," "$CONFIG_FILE"
fi

# 创建日志目录
mkdir -p "$LOG_DIR"
chown -R $(ps -o user= -p $TEA_PID):$(ps -o group= -p $TEA_PID) "$LOG_DIR" 2>/dev/null || true

echo "✅ 调试模式已启用"

# 重启服务
echo ""
echo "5. 重启Tea API服务..."
if systemctl is-active --quiet tea-api; then
    systemctl restart tea-api
    sleep 3
    if systemctl is-active --quiet tea-api; then
        echo "✅ 服务重启成功"
    else
        echo "❌ 服务重启失败"
        systemctl status tea-api
    fi
else
    echo "⚠️  无法通过systemctl重启，请手动重启Tea API服务"
    echo "您可以尝试："
    echo "  kill $TEA_PID"
    echo "  然后重新启动Tea API"
fi

# 创建日志监控脚本
echo ""
echo "6. 创建日志监控脚本..."
cat > /tmp/monitor_tea_logs.sh << 'EOF'
#!/bin/bash

echo "=== Tea API 日志监控 ==="
echo "监控空回复相关的错误..."
echo "按 Ctrl+C 停止监控"
echo ""

LOG_DIR="/var/log/tea-api"
JOURNAL_LOG=false

# 检查日志文件
if [ -d "$LOG_DIR" ] && [ "$(ls -A $LOG_DIR 2>/dev/null)" ]; then
    echo "监控日志目录: $LOG_DIR"
    tail -f $LOG_DIR/*.log 2>/dev/null | grep -E "(错误|失败|空|empty|invalid|error|Error|ERROR)" --color=always
elif journalctl -u tea-api --no-pager -n 1 >/dev/null 2>&1; then
    echo "监控systemd日志..."
    JOURNAL_LOG=true
    journalctl -u tea-api -f | grep -E "(错误|失败|空|empty|invalid|error|Error|ERROR)" --color=always
else
    echo "尝试监控所有可能的日志位置..."
    # 监控可能的日志位置
    (
        tail -f /var/log/tea-api/*.log 2>/dev/null &
        tail -f /opt/tea-api/*.log 2>/dev/null &
        tail -f ./oneapi*.log 2>/dev/null &
        journalctl -f 2>/dev/null | grep tea-api &
        wait
    ) | grep -E "(错误|失败|空|empty|invalid|error|Error|ERROR)" --color=always
fi
EOF

chmod +x /tmp/monitor_tea_logs.sh

echo "✅ 日志监控脚本已创建: /tmp/monitor_tea_logs.sh"

# 创建测试脚本
echo ""
echo "7. 创建API测试脚本..."
cat > /tmp/test_tea_api.sh << 'EOF'
#!/bin/bash

echo "=== Tea API 测试脚本 ==="
echo "此脚本将发送测试请求来重现空回复问题"
echo ""

# 获取API配置
read -p "请输入Tea API地址 (例如: http://localhost:3000): " API_URL
read -p "请输入API密钥: " API_KEY

if [ -z "$API_URL" ] || [ -z "$API_KEY" ]; then
    echo "❌ API地址和密钥不能为空"
    exit 1
fi

echo ""
echo "发送测试请求..."

# 测试聊天完成API
curl -X POST "$API_URL/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {
        "role": "user",
        "content": "你好，请回复一个简单的问候"
      }
    ],
    "max_tokens": 100,
    "temperature": 0.7
  }' \
  -v \
  -w "\n\n=== 响应统计 ===\nHTTP状态码: %{http_code}\n响应时间: %{time_total}s\n响应大小: %{size_download} bytes\n" \
  2>&1 | tee /tmp/tea_api_test_result.txt

echo ""
echo "测试结果已保存到: /tmp/tea_api_test_result.txt"
echo ""
echo "如果出现空回复，请检查以下内容："
echo "1. HTTP状态码是否为200"
echo "2. 响应大小是否为0或很小"
echo "3. 响应内容是否包含错误信息"
EOF

chmod +x /tmp/test_tea_api.sh

echo "✅ API测试脚本已创建: /tmp/test_tea_api.sh"

# 显示使用说明
echo ""
echo "=== 调试步骤说明 ==="
echo ""
echo "1. 启动日志监控："
echo "   /tmp/monitor_tea_logs.sh"
echo ""
echo "2. 在另一个终端运行API测试："
echo "   /tmp/test_tea_api.sh"
echo ""
echo "3. 观察日志输出，查找以下关键信息："
echo "   - '响应体为空' - 表示上游返回空响应"
echo "   - '解析响应体失败' - 表示响应格式有问题"
echo "   - '上游请求失败' - 表示网络连接问题"
echo "   - '流式响应无效' - 表示流式处理问题"
echo ""
echo "4. 常见解决方案："
echo "   - 检查渠道配置是否正确"
echo "   - 检查API密钥是否有效"
echo "   - 检查网络连接是否正常"
echo "   - 检查上游服务是否可用"
echo ""
echo "5. 恢复原配置（如果需要）："
echo "   cp ${CONFIG_FILE}.backup.* $CONFIG_FILE"
echo "   systemctl restart tea-api"
echo ""
echo "=== 调试完成 ==="
echo "现在您可以开始监控日志并测试API了"
