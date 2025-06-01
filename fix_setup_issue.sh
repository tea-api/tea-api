#!/bin/bash

# Tea API Setup 问题修复脚本
# 用于修复一直重定向到 /setup 页面的问题

echo "=== Tea API Setup 问题修复工具 ==="
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

# 确认操作
echo "此脚本将尝试修复 Tea API 的 Setup 问题。"
echo "请确保您已经备份了重要数据。"
echo ""
read -p "是否继续？(y/N): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "操作已取消。"
    exit 1
fi

# 1. 停止服务
echo "1. 停止 Tea API 服务..."
if command -v systemctl >/dev/null 2>&1; then
    if systemctl is-active --quiet tea-api; then
        echo "   正在停止服务..."
        sudo systemctl stop tea-api
        sleep 2
        if systemctl is-active --quiet tea-api; then
            print_error "服务停止失败"
            exit 1
        else
            print_status "服务已停止"
        fi
    else
        print_info "服务未运行"
    fi
else
    print_warning "systemctl 不可用，请手动停止服务"
fi
echo ""

# 2. 检查和修复数据库
echo "2. 检查和修复数据库..."
if [ -f "./tea-api.db" ]; then
    print_status "找到数据库文件: ./tea-api.db"
    
    # 备份数据库
    BACKUP_FILE="./tea-api.db.backup.$(date +%Y%m%d_%H%M%S)"
    echo "   创建备份: $BACKUP_FILE"
    cp "./tea-api.db" "$BACKUP_FILE"
    
    if command -v sqlite3 >/dev/null 2>&1; then
        echo "   检查数据库完整性..."
        if sqlite3 tea-api.db "PRAGMA integrity_check;" | grep -q "ok"; then
            print_status "数据库完整性检查通过"
        else
            print_error "数据库完整性检查失败"
            echo "   尝试从备份恢复或重新初始化"
        fi
        
        # 检查 Setup 表
        echo "   检查 Setup 表..."
        SETUP_COUNT=$(sqlite3 tea-api.db "SELECT COUNT(*) FROM setups;" 2>/dev/null || echo "0")
        ROOT_COUNT=$(sqlite3 tea-api.db "SELECT COUNT(*) FROM users WHERE role >= 100;" 2>/dev/null || echo "0")
        
        echo "   Setup 记录数: $SETUP_COUNT"
        echo "   Root 用户数: $ROOT_COUNT"
        
        if [ "$ROOT_COUNT" -gt 0 ] && [ "$SETUP_COUNT" -eq 0 ]; then
            print_warning "发现 Root 用户但没有 Setup 记录，正在修复..."
            
            # 获取当前时间戳
            CURRENT_TIME=$(date +%s)
            
            # 插入 Setup 记录
            sqlite3 tea-api.db "INSERT INTO setups (version, initialized_at) VALUES ('v0.0.0', $CURRENT_TIME);" 2>/dev/null
            
            if [ $? -eq 0 ]; then
                print_status "Setup 记录已创建"
            else
                print_error "创建 Setup 记录失败"
            fi
        elif [ "$ROOT_COUNT" -eq 0 ]; then
            print_warning "没有 Root 用户，系统需要重新初始化"
            echo "   删除现有 Setup 记录..."
            sqlite3 tea-api.db "DELETE FROM setups;" 2>/dev/null
        fi
    else
        print_warning "sqlite3 不可用，无法检查数据库内容"
    fi
else
    print_warning "数据库文件不存在，将在启动时自动创建"
fi
echo ""

# 3. 检查文件权限
echo "3. 检查和修复文件权限..."
if [ -f "./tea-api" ]; then
    if [ ! -x "./tea-api" ]; then
        echo "   修复可执行文件权限..."
        chmod +x ./tea-api
        print_status "可执行文件权限已修复"
    else
        print_status "可执行文件权限正常"
    fi
else
    print_error "tea-api 可执行文件不存在"
    exit 1
fi

if [ -f "./tea-api.db" ]; then
    echo "   修复数据库文件权限..."
    chmod 644 ./tea-api.db
    chown $(whoami):$(whoami) ./tea-api.db 2>/dev/null || true
    print_status "数据库文件权限已修复"
fi

# 确保日志目录存在
if [ ! -d "./logs" ]; then
    echo "   创建日志目录..."
    mkdir -p ./logs
    print_status "日志目录已创建"
fi
echo ""

# 4. 清理缓存和临时文件
echo "4. 清理缓存和临时文件..."
if [ -d "./tiktoken_cache" ]; then
    echo "   清理 tiktoken 缓存..."
    rm -rf ./tiktoken_cache/*
    print_status "tiktoken 缓存已清理"
fi

# 清理可能的锁文件
if [ -f "./tea-api.db-wal" ]; then
    echo "   清理 WAL 文件..."
    rm -f ./tea-api.db-wal
fi

if [ -f "./tea-api.db-shm" ]; then
    echo "   清理 SHM 文件..."
    rm -f ./tea-api.db-shm
fi
echo ""

# 5. 重新加载 systemd 配置
echo "5. 重新加载 systemd 配置..."
if command -v systemctl >/dev/null 2>&1; then
    sudo systemctl daemon-reload
    print_status "systemd 配置已重新加载"
else
    print_warning "systemctl 不可用"
fi
echo ""

# 6. 启动服务
echo "6. 启动 Tea API 服务..."
if command -v systemctl >/dev/null 2>&1; then
    echo "   正在启动服务..."
    sudo systemctl start tea-api
    sleep 3
    
    if systemctl is-active --quiet tea-api; then
        print_status "服务启动成功"
        echo "   服务状态:"
        systemctl status tea-api --no-pager -l | head -5
    else
        print_error "服务启动失败"
        echo "   查看错误日志:"
        journalctl -u tea-api --no-pager -n 10
        exit 1
    fi
else
    print_warning "systemctl 不可用，请手动启动服务"
    echo "   手动启动命令: ./tea-api --port 3000 --log-dir ./logs"
fi
echo ""

# 7. 测试 API
echo "7. 测试 API 响应..."
sleep 2
if command -v curl >/dev/null 2>&1; then
    echo "   正在测试 /api/setup..."
    SETUP_RESPONSE=$(curl -s -w "\nHTTP_CODE:%{http_code}" http://localhost:3000/api/setup 2>/dev/null)
    HTTP_CODE=$(echo "$SETUP_RESPONSE" | grep "HTTP_CODE:" | cut -d: -f2)
    RESPONSE_BODY=$(echo "$SETUP_RESPONSE" | sed '/HTTP_CODE:/d')
    
    if [ "$HTTP_CODE" = "200" ]; then
        print_status "API 响应正常"
        echo "   响应内容: $RESPONSE_BODY"
        
        if echo "$RESPONSE_BODY" | grep -q '"status":true'; then
            print_status "Setup 状态: 已完成 ✓"
            echo ""
            print_info "修复完成！现在可以访问 http://localhost:3000"
        else
            print_warning "Setup 状态: 仍需要初始化"
            echo "   请访问 http://localhost:3000/setup 完成初始化"
        fi
    else
        print_error "API 响应异常 (HTTP $HTTP_CODE)"
        echo "   请检查服务日志: sudo journalctl -u tea-api -f"
    fi
else
    print_warning "curl 不可用，请手动测试"
    echo "   请访问 http://localhost:3000 检查状态"
fi

echo ""
echo "=== 修复完成 ==="
echo ""
echo "后续建议:"
echo "1. 如果问题仍然存在，请查看实时日志: sudo journalctl -u tea-api -f"
echo "2. 检查浏览器控制台是否有错误信息"
echo "3. 尝试清除浏览器缓存和 Cookie"
echo "4. 如果使用反向代理，请检查代理配置"
