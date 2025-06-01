# Tea API 更新脚本使用说明

本目录包含了用于更新 Tea API 的脚本，可以自动拉取最新代码、重新编译并重启服务，同时保留您的配置文件。

## 脚本说明

### 1. update.sh - 完整更新脚本
功能最全面的更新脚本，包含交互式确认和详细的选项。

**特性：**
- 自动备份配置文件（.env、数据库、数据目录、日志）
- Git pull 最新代码
- 重新编译前端和后端
- 恢复配置文件
- 重启服务
- 清理旧备份（保留最近5个）
- 支持命令行参数

**使用方法：**
```bash
# 交互式更新（推荐）
./bin/update.sh

# 强制更新（无需确认）
./bin/update.sh --force

# 跳过备份的强制更新
./bin/update.sh --force --skip-backup

# 查看帮助
./bin/update.sh --help
```

### 2. quick_update.sh - 快速更新脚本
简化版本，无需交互，适合自动化脚本或快速更新。

**特性：**
- 自动备份配置
- Git pull 最新代码
- 重新编译
- 恢复配置
- 重启服务
- 清理旧备份（保留最近3个）

**使用方法：**
```bash
./bin/quick_update.sh
```

## 更新流程

两个脚本都遵循相同的更新流程：

1. **检查环境** - 验证是否在正确的项目目录中，检查 Git 是否可用
2. **备份配置** - 备份现有的配置文件和数据
3. **停止服务** - 安全停止正在运行的 tea-api 服务
4. **拉取代码** - 从 Git 仓库拉取最新代码
5. **重新编译** - 编译前端和后端代码
6. **恢复配置** - 将备份的配置文件恢复到新的部署目录
7. **启动服务** - 重新启动 tea-api 服务
8. **清理备份** - 删除过旧的备份文件

## 备份说明

### 自动备份的文件
- `.env` - 环境配置文件
- `tea-api.db` - SQLite 数据库文件
- `data/` - 数据目录
- `logs/` - 日志目录

### 备份目录命名
备份目录使用时间戳命名：`config-backup-YYYYMMDD-HHMMSS`

### 备份清理策略
- `update.sh`: 保留最近 5 个备份
- `quick_update.sh`: 保留最近 3 个备份

## 注意事项

### 运行前提
1. 必须在 Tea API 项目根目录运行
2. 需要 Git 仓库环境
3. 需要 sudo 权限（用于重启系统服务）
4. 确保已安装必要的依赖（Node.js, Go）

### 安全提示
1. 脚本会自动备份配置，但建议定期手动备份重要数据
2. 如果有未提交的代码更改，脚本会提示是否 stash
3. 更新过程中服务会短暂停止

### 故障恢复
如果更新失败，可以手动恢复：
1. 找到最新的备份目录（`config-backup-*`）
2. 手动复制配置文件回 `tea-api-linux-deploy/` 目录
3. 重启服务：`sudo systemctl restart tea-api`

## 常用命令

### 检查服务状态
```bash
sudo systemctl status tea-api
```

### 查看服务日志
```bash
sudo journalctl -u tea-api -f
```

### 手动重启服务
```bash
sudo systemctl restart tea-api
```

### 手动启动应用
```bash
cd tea-api-linux-deploy
./start.sh
```

## 示例输出

更新成功后，您会看到类似的输出：
```
=== Update Complete ===
[INFO] Tea API has been updated successfully!
[INFO] Current version: v1.2.3
[INFO] Service status: Running
[INFO] Access your application at: http://localhost:3000
[INFO] Latest backup: config-backup-20240101-120000
```

## 故障排除

### 常见问题

1. **权限错误**
   - 确保脚本有执行权限：`chmod +x bin/update.sh`
   - 确保有 sudo 权限

2. **Git 错误**
   - 检查网络连接
   - 确保 Git 配置正确
   - 处理未提交的更改

3. **编译错误**
   - 检查 Node.js 和 Go 是否正确安装
   - 检查依赖是否完整

4. **服务启动失败**
   - 检查配置文件是否正确
   - 查看服务日志：`sudo journalctl -u tea-api -f`
   - 检查端口是否被占用

如果遇到问题，可以查看备份目录中的配置文件，手动恢复配置。
