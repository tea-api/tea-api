#!/bin/bash

# Tea API 初始化状态调试脚本

echo "=== Tea API 初始化状态调试 ==="
echo ""

# 检查容器状态
echo "1. 检查容器状态："
docker ps -a | grep tea-api
echo ""

# 检查容器日志
echo "2. 检查最新日志（最后20行）："
docker logs --tail 20 tea-api
echo ""

# 检查 /api/status 接口
echo "3. 检查 /api/status 接口："
curl -s http://localhost:3000/api/status | jq '.data.setup' 2>/dev/null || echo "无法访问 API 或解析 JSON"
echo ""

# 检查 /api/setup 接口
echo "4. 检查 /api/setup 接口："
curl -s http://localhost:3000/api/setup | jq '.' 2>/dev/null || echo "无法访问 API 或解析 JSON"
echo ""

# 提供解决方案
echo "=== 可能的解决方案 ==="
echo ""
echo "如果系统仍然显示需要初始化，请尝试以下步骤："
echo ""
echo "1. 重新构建 Docker 镜像："
echo "   ./rebuild-docker.sh"
echo ""
echo "2. 清除浏览器缓存："
echo "   - 按 Ctrl+Shift+R 强制刷新页面"
echo "   - 或者在开发者工具中清除缓存"
echo ""
echo "3. 检查数据库中的 setup 记录："
echo "   docker exec -it tea-api sqlite3 /app/data/tea-api.db \"SELECT * FROM setups;\""
echo ""
echo "4. 如果使用外部数据库，检查数据库连接："
echo "   docker logs tea-api | grep -i \"database\""
echo ""
echo "5. 手动访问初始化页面："
echo "   http://localhost:3000/setup"
echo ""
