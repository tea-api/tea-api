#!/bin/bash

# Tea API Setup 诊断脚本
# 用于诊断为什么系统一直重定向到 /setup 页面

echo "=== Tea API Setup 诊断工具 ==="
echo "时间: $(date)"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

# 1. 检查服务状态
echo "1. 检查 Tea API 服务状态:"
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet tea-api; then
        print_status "服务正在运行"
        echo "   服务状态详情:"
        systemctl status tea-api --no-pager -l | head -10
    else
        print_error "服务未运行"
        echo "   尝试启动服务: sudo systemctl start tea-api"
    fi
else
    print_warning "systemctl 不可用，无法检查服务状态"
fi
echo ""

# 2. 检查 API 响应
echo "2. 检查 Setup API 响应:"
if command -v curl >/dev/null 2>&1; then
    echo "   正在调用 /api/setup..."
    SETUP_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" http://localhost:3000/api/setup 2>/dev/null)
    HTTP_CODE=$(echo "$SETUP_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
    RESPONSE_BODY=$(echo "$SETUP_RESPONSE" | sed '/HTTP_CODE:/d')
    
    if [ "$HTTP_CODE" = "200" ]; then
        print_status "API 响应正常 (HTTP $HTTP_CODE)"
        echo "   响应内容: $RESPONSE_BODY"
        
        # 解析 JSON 响应
        if echo "$RESPONSE_BODY" | grep -q '"status":true'; then
            print_status "Setup 状态: 已完成"
        elif echo "$RESPONSE_BODY" | grep -q '"status":false'; then
            print_error "Setup 状态: 未完成"
        else
            print_warning "无法解析 Setup 状态"
        fi
    else
        print_error "API 响应异常 (HTTP $HTTP_CODE)"
        echo "   响应内容: $RESPONSE_BODY"
    fi
else
    print_warning "curl 不可用，无法测试 API"
fi
echo ""

# 3. 检查数据库
echo "3. 检查数据库状态:"
if [ -f "./tea-api" ]; then
    # 检查数据库文件
    if [ -f "./tea-api.db" ]; then
        print_status "SQLite 数据库文件存在: ./tea-api.db"
        echo "   文件大小: $(ls -lh tea-api.db | awk '{print $5}')"
        echo "   修改时间: $(ls -l tea-api.db | awk '{print $6, $7, $8}')"
        
        # 检查数据库内容
        if command -v sqlite3 >/dev/null 2>&1; then
            echo "   检查 Setup 表:"
            SETUP_COUNT=$(sqlite3 tea-api.db "SELECT COUNT(*) FROM setups;" 2>/dev/null || echo "ERROR")
            if [ "$SETUP_COUNT" = "ERROR" ]; then
                print_error "无法查询 Setup 表（表可能不存在）"
            elif [ "$SETUP_COUNT" -gt 0 ]; then
                print_status "Setup 记录存在 ($SETUP_COUNT 条)"
                echo "   Setup 记录详情:"
                sqlite3 tea-api.db "SELECT id, version, datetime(initialized_at, 'unixepoch', 'localtime') as init_time FROM setups;" 2>/dev/null || echo "   查询失败"
            else
                print_error "Setup 表为空"
            fi
            
            echo "   检查用户表:"
            ROOT_COUNT=$(sqlite3 tea-api.db "SELECT COUNT(*) FROM users WHERE role >= 100;" 2>/dev/null || echo "ERROR")
            if [ "$ROOT_COUNT" = "ERROR" ]; then
                print_error "无法查询用户表"
            elif [ "$ROOT_COUNT" -gt 0 ]; then
                print_status "Root 用户存在 ($ROOT_COUNT 个)"
            else
                print_error "没有 Root 用户"
            fi
        else
            print_warning "sqlite3 不可用，无法检查数据库内容"
        fi
    else
        print_error "SQLite 数据库文件不存在"
        echo "   当前目录: $(pwd)"
        echo "   目录内容: $(ls -la | grep -E '\.(db|sqlite)$' || echo '无数据库文件')"
    fi
else
    print_error "tea-api 可执行文件不存在"
fi
echo ""

# 4. 检查环境变量
echo "4. 检查环境变量:"
if [ -f ".env" ]; then
    print_status ".env 文件存在"
    echo "   相关配置:"
    grep -E "^(SQL_DSN|SESSION_SECRET|GIN_MODE)" .env 2>/dev/null || echo "   无相关配置"
else
    print_warning ".env 文件不存在"
fi

# 检查系统环境变量
if [ -n "$SQL_DSN" ]; then
    echo "   SQL_DSN: $SQL_DSN"
else
    echo "   SQL_DSN: 未设置（将使用 SQLite）"
fi
echo ""

# 5. 检查日志
echo "5. 检查最近的日志:"
if command -v journalctl >/dev/null 2>&1; then
    echo "   系统日志 (最近 10 条):"
    journalctl -u tea-api --no-pager -n 10 --since "10 minutes ago" 2>/dev/null | grep -E "(setup|Setup|SETUP|initialized)" || echo "   无相关日志"
else
    print_warning "journalctl 不可用"
fi

if [ -d "./logs" ]; then
    echo "   应用日志:"
    find ./logs -name "*.log" -mtime -1 -exec echo "   文件: {}" \; -exec grep -l -E "(setup|Setup|SETUP|initialized)" {} \; 2>/dev/null | head -5
else
    print_warning "logs 目录不存在"
fi
echo ""

# 6. 诊断建议
echo "6. 诊断建议:"
echo ""

# 检查常见问题
ISSUES_FOUND=0

# 检查服务是否运行
if ! systemctl is-active --quiet tea-api 2>/dev/null; then
    print_error "问题: 服务未运行"
    echo "   解决方案: sudo systemctl start tea-api"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

# 检查数据库
if [ ! -f "./tea-api.db" ] && [ -z "$SQL_DSN" ]; then
    print_error "问题: 数据库文件不存在且未配置外部数据库"
    echo "   解决方案: 确保应用有权限创建数据库文件"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

# 检查权限
if [ -f "./tea-api.db" ] && [ ! -w "./tea-api.db" ]; then
    print_error "问题: 数据库文件无写权限"
    echo "   解决方案: sudo chown \$(whoami):\$(whoami) tea-api.db"
    ISSUES_FOUND=$((ISSUES_FOUND + 1))
fi

if [ $ISSUES_FOUND -eq 0 ]; then
    print_info "未发现明显问题，可能需要进一步调试"
    echo ""
    echo "建议的调试步骤:"
    echo "1. 重启服务: sudo systemctl restart tea-api"
    echo "2. 查看实时日志: sudo journalctl -u tea-api -f"
    echo "3. 手动运行应用: ./tea-api --port 3001 --log-dir ./logs"
    echo "4. 检查浏览器控制台错误"
    echo "5. 清除浏览器缓存和 Cookie"
fi

echo ""
echo "=== 诊断完成 ==="
