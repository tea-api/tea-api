# Tea API 初始化问题解决方案

## 问题描述

您遇到的问题是：虽然系统已经初始化过，但前端仍然显示需要初始化的页面。

## 问题原因分析

1. **数据库状态与内存状态不同步**：数据库中存在 setup 记录，但 `constant.Setup` 变量可能没有正确设置为 `true`
2. **前端缓存问题**：浏览器缓存了旧的状态信息
3. **Docker 容器重启后状态丢失**：容器重启时可能没有正确读取数据库中的 setup 状态

## 解决方案

### 1. 代码修复

我已经对代码进行了以下修改：

#### `model/main.go` 中的 `CheckSetup()` 函数
- 添加了更详细的日志输出
- 改进了错误处理逻辑
- 添加了状态确认日志

#### `controller/misc.go` 中的 `GetStatus()` 函数
- 添加了调试日志，显示 `constant.Setup` 的当前值

### 2. 重新构建和部署

使用提供的脚本重新构建 Docker 镜像：

```bash
# 1. 重新构建镜像
./rebuild-docker.sh

# 2. 启动容器（选择适合您的数据库配置）
# SQLite 版本：
docker run -d --name tea-api -p 3000:3000 -v $(pwd)/data:/data tea-api:latest

# MySQL 版本：
docker run -d --name tea-api -p 3000:3000 \
  -e SQL_DSN='user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local' \
  -v $(pwd)/data:/data \
  tea-api:latest

# PostgreSQL 版本：
docker run -d --name tea-api -p 3000:3000 \
  -e SQL_DSN='postgres://user:password@host:port/dbname?sslmode=disable' \
  -v $(pwd)/data:/data \
  tea-api:latest
```

### 3. 调试和验证

使用调试脚本检查状态：

```bash
./debug-setup.sh
```

### 4. 手动解决步骤

如果问题仍然存在，请按以下步骤操作：

#### 步骤 1：检查日志
```bash
docker logs -f tea-api
```

查找以下关键日志：
- `system is already initialized at: ...`
- `CheckSetup completed: constant.Setup = true`
- `GetStatus: constant.Setup = true`

#### 步骤 2：清除浏览器缓存
- 按 `Ctrl+Shift+R` 强制刷新页面
- 或在开发者工具中清除缓存和硬重载

#### 步骤 3：检查 API 响应
```bash
# 检查状态接口
curl http://localhost:3000/api/status | jq '.data.setup'

# 检查初始化接口
curl http://localhost:3000/api/setup | jq '.'
```

#### 步骤 4：数据库检查（SQLite）
```bash
# 进入容器
docker exec -it tea-api sh

# 检查 setup 表
sqlite3 /data/tea-api.db "SELECT * FROM setups;"
```

#### 步骤 5：数据库检查（MySQL/PostgreSQL）
使用相应的数据库客户端连接并执行：
```sql
SELECT * FROM setups;
```

### 5. 应急解决方案

如果以上方法都不能解决问题，可以尝试以下应急方案：

#### 方案 A：重置初始化状态
```bash
# 删除 setup 记录，让系统重新初始化
docker exec -it tea-api sqlite3 /data/tea-api.db "DELETE FROM setups;"
```

#### 方案 B：手动设置初始化状态
```bash
# 手动插入 setup 记录
docker exec -it tea-api sqlite3 /data/tea-api.db "INSERT INTO setups (version, initialized_at) VALUES ('v0.0.0', $(date +%s));"
```

## 预防措施

为了避免将来出现类似问题：

1. **定期备份数据库**
2. **使用持久化存储**：确保 Docker 容器的数据目录正确挂载
3. **监控日志**：定期检查应用日志，及时发现问题
4. **版本控制**：记录每次部署的版本和配置

## 联系支持

如果问题仍然无法解决，请提供以下信息：

1. Docker 容器日志（`docker logs tea-api`）
2. API 响应（`curl http://localhost:3000/api/status`）
3. 数据库配置（脱敏后）
4. 浏览器开发者工具中的网络请求信息

## 更新日志

- 2025-06-01: 修复了 CheckSetup 函数的日志输出和错误处理
- 2025-06-01: 添加了 GetStatus 接口的调试日志
- 2025-06-01: 创建了自动化调试和重建脚本
