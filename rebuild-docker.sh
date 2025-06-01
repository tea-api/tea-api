#!/bin/bash

# Tea API Docker 重新构建脚本
# 用于解决初始化状态问题

echo "开始重新构建 Tea API Docker 镜像..."

# 停止现有容器
echo "停止现有容器..."
docker stop tea-api 2>/dev/null || true
docker rm tea-api 2>/dev/null || true

# 构建新镜像
echo "构建新镜像..."
docker build -t tea-api:latest .

if [ $? -ne 0 ]; then
    echo "Docker 镜像构建失败！"
    exit 1
fi

echo "Docker 镜像构建成功！"

# 提供启动命令示例
echo ""
echo "请使用以下命令启动容器："
echo ""
echo "# 基本启动命令（使用 SQLite）："
echo "docker run -d --name tea-api -p 3000:3000 -v \$(pwd)/data:/app/data tea-api:latest"
echo ""
echo "# 使用 MySQL 数据库："
echo "docker run -d --name tea-api -p 3000:3000 \\"
echo "  -e SQL_DSN='user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local' \\"
echo "  -v \$(pwd)/data:/app/data \\"
echo "  tea-api:latest"
echo ""
echo "# 使用 PostgreSQL 数据库："
echo "docker run -d --name tea-api -p 3000:3000 \\"
echo "  -e SQL_DSN='postgres://user:password@host:port/dbname?sslmode=disable' \\"
echo "  -v \$(pwd)/data:/app/data \\"
echo "  tea-api:latest"
echo ""
echo "启动后请检查日志："
echo "docker logs -f tea-api"
