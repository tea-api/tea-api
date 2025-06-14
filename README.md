<p align="right">
  <strong><a href="./README.en.md">English</a> | 中文</strong>
</p>

<a href="https://github.com/tea-api/tea-api" target="_blank">
  <img src="/web/public/logo.png" alt="tea-api" width="100" height="100" />
</a>

# Tea-API

[![License](https://img.shields.io/github/license/tea-api/tea-api)](https://github.com/tea-api/tea-api/blob/main/LICENSE) [![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tea-api/tea-api)](https://github.com/tea-api/tea-api/releases)

API公益站系统

基于原汁原味的 New API 体验, 对界面无大改动, 遵循 Apache 2.0 协议, 无商用限制, 承诺不变质.  
添加极多原版不计划添加的特性. 以下只是部分.

## 目录

- [特性](#特性)
- [迁移](#迁移)
- [部署](#部署)
- [环境变量](#环境变量)
- [Star History](#-star-history)
- [License](#license)

## 特性

- 🚀 支持原版new-api所有功能
- 🔄 多渠道负载均衡：智能分配请求，确保高可用性
- 📊 完善的用户管理：用户注册、登录、Token管理、使用统计
- 💰 内置计费系统：支持充值、兑换码、签到奖励等功能
- 🛡️ 安全防护：速率限制、IP限制、防刷等安全措施
- 🔌 超级优化支持：提升系统性能和响应速度
- 🐳 完整的Docker支持：快速部署，简化运维

## 迁移

本程序基于 new-api 二开, 数据库结构基本兼容, 会自动运行迁移.  
其他类似程序不保证支持, 后续有计划做手动迁移指南.  

### new-api

除了使用 SQLite, 均可无缝迁移.  
对于 SQLite, 建议将 `one-api.db` 重命名为 `tea-api.db`, 系统会尝试自动处理, 但未经过测试. 

## 部署

> [!TIP]
> 最新版 Docker 镜像：`ghcr.io/teapi/tea-api:latest`

### 快速部署 (Linux)

对于 Linux (Debian/Ubuntu) 系统，我们提供了一键部署脚本：

```bash
# 克隆项目
git clone https://github.com/tea-api/tea-api.git
cd tea-api

# 一键部署（自动安装依赖、构建、配置、启动）
./deploy_linux.sh
```

### docker-compose

1. 克隆此仓库

```shell
git clone https://github.com/tea-api/tea-api.git
cd tea-api
```

2. 修改配置文件

```shell
nano docker-compose.yml
```

3. 构建并启动服务

```shell
docker-compose up -d --build
```

如需单独构建镜像，可执行：

```shell
docker build -t teapi/tea-api:latest .
```

### 手动部署

#### Linux 系统

1. 设置环境：
```bash
./bin/setup_env_linux.sh
```

2. 构建应用：
```bash
./bin/build_linux.sh
```

3. 配置并启动：
```bash
cd tea-api-linux-deploy
cp .env.example .env
# 编辑 .env 文件配置数据库等
./start.sh
```

#### macOS 系统

1. 设置环境：
```bash
./bin/setup_env_mac.sh
```

2. 构建应用：
```bash
./bin/build_mac.sh
```

3. 运行：
```bash
./tea-api-macos --port 3000 --log-dir ./logs
```

#### 使用 Makefile

```bash
# 设置 Linux 环境
make setup-linux

# 构建 Linux 版本
make build-linux

# 设置 macOS 环境
make setup-mac

# 构建 macOS 版本
make build-mac

# 快速 Linux 部署
make deploy-linux

# 清理构建文件
make clean
```

更多部署信息请参考 [Linux 部署指南](docs/installation/linux-debian.md)。

## 环境变量

| 环境变量 | 说明 | 默认值 |
|---------|------|--------|
| `SQL_DSN` | 数据库连接字符串 | `./tea-api.db` |
| `REDIS_CONN_STRING` | Redis连接字符串 | - |
| `TZ` | 时区设置 | `Asia/Shanghai` |
| `ERROR_LOG_ENABLED` | 是否启用错误日志 | `false` |
| `TIKTOKEN_CACHE_DIR` | tiktoken缓存目录 | `./tiktoken_cache` |
| `SESSION_SECRET` | 会话密钥(多机部署必须) | 随机字符串 |
| `CRYPTO_SECRET` | 加密密钥 | 随机字符串 |
| `NODE_TYPE` | 节点类型(master/slave) | `master` |
| `SYNC_FREQUENCY` | 数据同步频率(秒) | `60` |
| `FRONTEND_BASE_URL` | 前端基础URL | - |
| `MEMORY_CACHE_ENABLED` | 启用内存缓存 | `true` |
| `RATE_LIMIT_ENABLED` | 启用速率限制 | `true` |
| `RATE_LIMIT_REDIS` | 速率限制Redis连接 | 同`REDIS_CONN_STRING` |

## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tea-api/tea-api&type=Date)](https://star-history.com/#tea-api/tea-api&Date)

## License

Tea-API is released under the [Apache-2.0](LICENSE) License.

</div>
