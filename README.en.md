<p align="right">
  <strong><a href="./README.md">中文</a> | English</strong>
</p>
<div align="center">

![tea-api](/web/public/logo.png)

# Tea-API

[![License](https://img.shields.io/github/license/tea-api/tea-api)](https://github.com/tea-api/tea-api/blob/main/LICENSE) [![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tea-api/tea-api)](https://github.com/tea-api/tea-api/releases)

API公益站系统


基于原汁原味的 New API 体验, 对界面无大改动, 遵循 Apache 2.0 协议, 无商用限制, 承诺不变质.  
添加极多原版不计划添加的特性. 以下只是部分.

## Table of Contents

- [Features](#特性)
- [Migration](#迁移)
- [Deployment](#部署)
- [Environment Variables](#环境变量)
- [Star History](#-star-history)
- [License](#license)

## 特性


- 🚀 Full compatibility with the original new-api
- 🔄 Multi-channel load balancing for high availability
- 📊 Complete user management: sign up, login, token statistics
- 💰 Built-in billing system with recharge, redeem code and check-in rewards
- 🛡️ Security features including rate limiting and IP restrictions
- 🔌 Performance optimizations for faster response
- 🐳 Comprehensive Docker support for easy deployment

## 迁移

本程序基于 new-api 二开, 数据库结构基本兼容, 会自动运行迁移.  
其他类似程序不保证支持, 后续有计划做手动迁移指南.  

### new-api

除了使用 SQLite, 均可无缝迁移.  
对于 SQLite, 建议将 `one-api.db` 重命名为 `tea-api.db`, 系统会尝试自动处理, 但未经过测试. 

## 部署

> [!TIP]
> 最新版 Docker 镜像：`ghcr.io/teapi/tea-api:latest`

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

3. 启动服务

```shell
docker-compose up -d
```

## 环境变量

| Variable | Description | Default |
|----------|-------------|---------|
| `SQL_DSN` | Database connection string | `./tea-api.db` |
| `REDIS_CONN_STRING` | Redis connection | - |
| `TZ` | Time zone | `Asia/Shanghai` |
| `ERROR_LOG_ENABLED` | Enable error log | `false` |
| `TIKTOKEN_CACHE_DIR` | tiktoken cache dir | `./tiktoken_cache` |
| `SESSION_SECRET` | Session secret (required for multi-node) | random |
| `CRYPTO_SECRET` | Crypto secret | random |
| `NODE_TYPE` | Node type (master/slave) | `master` |
| `SYNC_FREQUENCY` | Sync frequency (seconds) | `60` |
| `FRONTEND_BASE_URL` | Frontend base URL | - |
| `MEMORY_CACHE_ENABLED` | Enable memory cache | `true` |
| `RATE_LIMIT_ENABLED` | Enable rate limit | `true` |
| `RATE_LIMIT_REDIS` | Rate limit Redis connection | same as `REDIS_CONN_STRING` |
## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=tea-api/tea-api&type=Date)](https://star-history.com/#tea-api/tea-api&Date)

## License

Tea-API is released under the [Apache-2.0](LICENSE) License.

</div>
