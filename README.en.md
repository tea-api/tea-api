<p align="right">
  <strong><a href="./README.md">中文</a> | English</strong>
</p>

![tea-api](/web/public/logo.png)

# Tea-API

[![License](https://img.shields.io/github/license/tea-api/tea-api)](https://github.com/tea-api/tea-api/blob/main/LICENSE) [![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tea-api/tea-api)](https://github.com/tea-api/tea-api/releases)

API Public Station System

Based on the authentic New API experience, there are no major changes to the interface, adhering to the Apache 2.0 license, with no commercial restrictions, and a commitment to remain unchanged.  
Many features that were not planned to be added in the original version have been included. The following is just a part.

## Table of Contents

- [Features](#Features)
- [Migration](#Migration)
- [Deployment](#Deployment)
- [Environment Variables](#Environment Variables)
- [Star History](#-star-history)
- [License](#license)

## Features


- 🚀 Full compatibility with the original new-api
- 🔄 Multi-channel load balancing for high availability
- 📊 Complete user management: sign up, login, token statistics
- 💰 Built-in billing system with recharge, redeem code and check-in rewards
- 🛡️ Security features including rate limiting and IP restrictions
- 🔌 Performance optimizations for faster response
- 🐳 Comprehensive Docker support for easy deployment

## Migration

This program is based on the second development of new-api, and the database structure is basically compatible, which will run migration automatically.  
Other similar programs do not guarantee support, and there are plans to create a manual migration guide in the future.  

### new-api

Seamless migration is possible except for using SQLite.  
For SQLite, it is recommended to rename `one-api.db` to `tea-api.db`, and the system will attempt to handle it automatically, but it has not been tested. 

## Deployment

> [!TIP]
> Latest Docker image: `ghcr.io/teapi/tea-api:latest`

### docker-compose

1. Clone this repository

```shell
git clone https://github.com/tea-api/tea-api.git
cd tea-api
```

2. Modify the configuration file

```shell
nano docker-compose.yml
```

3. Start the service

```shell
docker-compose up -d
```

## Environment Variables

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
