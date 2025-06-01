# Linux (Debian/Ubuntu) 安装指南

本指南将帮助您在 Linux Debian/Ubuntu 系统上安装和运行 Tea API。

## 系统要求

- **操作系统**: Ubuntu 18.04+ 或 Debian 10+
- **架构**: x86_64 (amd64) 或 ARM64
- **内存**: 最少 512MB，推荐 1GB+
- **存储**: 最少 1GB 可用空间
- **网络**: 需要互联网连接下载依赖

## 快速部署

如果您想要一键部署，可以使用我们的快速部署脚本：

```bash
# 克隆项目
git clone https://github.com/your-repo/tea-api.git
cd tea-api

# 运行快速部署脚本
./deploy_linux.sh
```

这个脚本会自动：
1. 安装所有必需的依赖
2. 构建前端和后端
3. 配置应用程序
4. 部署为系统服务

## 手动安装

### 1. 安装依赖

#### 更新系统包
```bash
sudo apt update && sudo apt upgrade -y
```

#### 安装基础工具
```bash
sudo apt install -y build-essential curl wget git unzip ca-certificates
```

#### 安装 Node.js 18.x
```bash
curl -fsSL https://deb.nodesource.com/setup_18.x | sudo -E bash -
sudo apt-get install -y nodejs
```

#### 安装 Go 1.23.4
```bash
# 下载 Go
wget https://go.dev/dl/go1.23.4.linux-amd64.tar.gz

# 安装 Go
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.23.4.linux-amd64.tar.gz

# 添加到 PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# 验证安装
go version
```

### 2. 克隆项目

```bash
git clone https://github.com/your-repo/tea-api.git
cd tea-api
```

### 3. 构建应用

#### 使用构建脚本（推荐）
```bash
# 设置环境
./bin/setup_env_linux.sh

# 构建应用
./bin/build_linux.sh
```

#### 手动构建
```bash
# 构建前端
cd web
npm install
DISABLE_ESLINT_PLUGIN='true' VITE_REACT_APP_VERSION=$(cat ../VERSION) npm run build
cd ..

# 构建后端
go mod download
CGO_ENABLED=0 GOOS=linux go build -ldflags "-s -w -X 'tea-api/common.Version=$(cat VERSION)' -extldflags '-static'" -o tea-api
```

### 4. 配置应用

#### 创建配置文件
```bash
cp .env.example .env
```

#### 编辑配置文件
```bash
nano .env
```

基本配置示例：
```env
# 数据库配置
SQL_DSN=./tea-api.db

# 服务器配置
TZ=Asia/Shanghai
ERROR_LOG_ENABLED=true
TIKTOKEN_CACHE_DIR=./tiktoken_cache

# 安全配置（生产环境必须设置）
SESSION_SECRET=your_random_session_secret_here
CRYPTO_SECRET=your_random_crypto_secret_here

# 缓存配置
MEMORY_CACHE_ENABLED=true

# 速率限制
RATE_LIMIT_ENABLED=true
```

#### 创建必要目录
```bash
mkdir -p data logs tiktoken_cache
```

## 部署方式

### 方式一：系统服务（推荐）

#### 1. 创建服务文件
```bash
sudo cp tea-api.service /etc/systemd/system/
```

#### 2. 修改服务文件
```bash
sudo nano /etc/systemd/system/tea-api.service
```

更新路径和用户：
```ini
[Unit]
Description=Tea API Service
After=network.target

[Service]
User=your_username
WorkingDirectory=/path/to/tea-api
ExecStart=/path/to/tea-api/tea-api --port 3000 --log-dir /path/to/tea-api/logs
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

#### 3. 启动服务
```bash
sudo systemctl daemon-reload
sudo systemctl enable tea-api
sudo systemctl start tea-api
```

#### 4. 检查状态
```bash
sudo systemctl status tea-api
sudo journalctl -u tea-api -f
```

### 方式二：手动启动

```bash
# 前台运行
./tea-api --port 3000 --log-dir ./logs

# 后台运行
nohup ./tea-api --port 3000 --log-dir ./logs > /dev/null 2>&1 &
```

### 方式三：Docker 部署

#### 1. 安装 Docker
```bash
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER
```

#### 2. 使用 Docker Compose
```bash
docker compose up -d
```

## 数据库配置

### SQLite（默认）
```env
SQL_DSN=./tea-api.db
```

### MySQL
```env
SQL_DSN=username:password@tcp(localhost:3306)/database_name
```

### PostgreSQL
```env
SQL_DSN=postgres://username:password@localhost:5432/database_name
```

## Redis 配置（可选）

```env
REDIS_CONN_STRING=redis://localhost:6379
# 或带密码
REDIS_CONN_STRING=redis://:password@localhost:6379
```

## 防火墙配置

### UFW
```bash
sudo ufw allow 3000/tcp
sudo ufw reload
```

### iptables
```bash
sudo iptables -A INPUT -p tcp --dport 3000 -j ACCEPT
sudo iptables-save > /etc/iptables/rules.v4
```

## SSL/TLS 配置

### 使用 Nginx 反向代理

#### 1. 安装 Nginx
```bash
sudo apt install nginx
```

#### 2. 配置 Nginx
```bash
sudo nano /etc/nginx/sites-available/tea-api
```

```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

#### 3. 启用站点
```bash
sudo ln -s /etc/nginx/sites-available/tea-api /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

#### 4. 安装 SSL 证书（Let's Encrypt）
```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## 故障排除

### 常见问题

#### 1. 端口被占用
```bash
# 查看端口占用
sudo netstat -tlnp | grep :3000

# 杀死进程
sudo kill -9 <PID>
```

#### 2. 权限问题
```bash
# 确保文件权限正确
chmod +x tea-api
chown -R $USER:$USER .
```

#### 3. 依赖问题
```bash
# 重新安装依赖
cd web && rm -rf node_modules && npm install
go mod download
```

#### 4. 数据库连接问题
```bash
# 检查数据库服务状态
sudo systemctl status mysql
sudo systemctl status postgresql
```

### 日志查看

#### 应用日志
```bash
tail -f logs/oneapi-*.log
```

#### 系统服务日志
```bash
sudo journalctl -u tea-api -f
```

#### Nginx 日志
```bash
sudo tail -f /var/log/nginx/access.log
sudo tail -f /var/log/nginx/error.log
```

## 性能优化

### 系统优化
```bash
# 增加文件描述符限制
echo "* soft nofile 65536" | sudo tee -a /etc/security/limits.conf
echo "* hard nofile 65536" | sudo tee -a /etc/security/limits.conf
```

### 应用优化
```env
# 启用内存缓存
MEMORY_CACHE_ENABLED=true

# 配置 Redis 缓存
REDIS_CONN_STRING=redis://localhost:6379
```

## 备份和恢复

### 数据备份
```bash
# SQLite 备份
cp tea-api.db tea-api.db.backup

# MySQL 备份
mysqldump -u username -p database_name > backup.sql

# PostgreSQL 备份
pg_dump -U username database_name > backup.sql
```

### 配置备份
```bash
tar -czf tea-api-backup.tar.gz .env logs data
```

## 更新升级

### 更新应用
```bash
# 停止服务
sudo systemctl stop tea-api

# 备份数据
cp tea-api.db tea-api.db.backup

# 拉取最新代码
git pull

# 重新构建
./bin/build_linux.sh

# 启动服务
sudo systemctl start tea-api
```

## 监控

### 系统监控
```bash
# 查看资源使用
htop
df -h
free -h
```

### 应用监控
```bash
# 查看进程状态
ps aux | grep tea-api

# 查看网络连接
ss -tlnp | grep :3000
```

## 支持

如果您遇到问题，请：

1. 查看日志文件
2. 检查系统资源
3. 验证配置文件
4. 查看 GitHub Issues
5. 联系技术支持

---

更多信息请参考：
- [API 文档](../api/)
- [配置说明](../configuration/)
- [故障排除](../troubleshooting/)
