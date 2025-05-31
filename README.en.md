<p align="right">
   <a href="./README.en.md">English</a> | <strong>中文</strong>
</p>
<div align="center">

![tea-api](/web/public/logo.png)

# Veloera

[![License](https://img.shields.io/github/license/tea-api/tea-api)](https://github.com/tea-api/tea-api/blob/main/LICENSE) [![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/tea-api/tea-api)](https://github.com/tea-api/tea-api/releases)

API公益站系统


基于原汁原味的 New API 体验, 对界面无大改动, 遵循 Apache 2.0 协议, 无商用限制, 承诺不变质.  
添加极多原版不计划添加的特性. 以下只是部分.  

## 特性



## 迁移

本程序基于 new-api 二开, 数据库结构基本兼容, 会自动运行迁移.  
其他类似程序不保证支持, 后续有计划做手动迁移指南.  

### new-api

除了使用 SQLite, 均可无缝迁移.  
对于 SQLite, 建议将 `one-api.db` 重命名为 `tea-api.db`, 系统会尝试自动处理, 但未经过测试. 

## 部署

> [!TIP]
> 最新版 Docker 镜像：`ghcr.io/veloera/veloera:latest`

### docker-compose

1. 克隆此仓库

```shell
git clone https://github.com/tea-api/tea-api.git
cd veloera
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



## 🌟 Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Veloera/Veloera&type=Date)](https://star-history.com/#Veloera/Veloera&Date)
