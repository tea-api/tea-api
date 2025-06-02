# Tea API 前端安全设置组件

## 概述

这是一套完整的前端安全设置管理界面，为 Tea API 提供可视化的安全配置和监控功能。

## 组件结构

```
web/src/components/Security/
├── SecurityOverview.js          # 安全概览页面
├── AbnormalDetectionSettings.js # 异常检测设置
├── RequestLimitSettings.js      # 请求限制设置
├── StreamProtectionSettings.js  # 流保护设置
├── IPBlacklistSettings.js       # IP黑名单管理
├── SecurityLogs.js              # 安全日志查看
└── README.md                    # 本文档
```

## 主要功能

### 1. 安全概览 (SecurityOverview.js)

- **安全评分显示**：基于各项安全指标计算的综合评分
- **实时统计**：显示被阻止的请求、恶意检测、黑名单IP等数据
- **防护状态**：各个安全模块的启用状态
- **可疑活动统计**：展示各类可疑活动的统计信息

### 2. 异常检测设置 (AbnormalDetectionSettings.js)

- **基础开关**：启用/禁用异常检测和自动黑名单
- **内容检测**：配置最大Prompt长度和随机字符比例阈值
- **频率检测**：设置最小请求间隔和可疑分数限制
- **流检测**：配置最大并发流数量和超时时间

### 3. 请求限制设置 (RequestLimitSettings.js)

- **请求体大小限制**：设置最大请求体大小
- **内容限制**：配置Prompt长度、消息数量等限制
- **内容质量验证**：启用随机内容和重复内容检测

### 4. 流保护设置 (StreamProtectionSettings.js)

- **并发限制**：设置每个IP和用户的最大流数量
- **超时配置**：配置流空闲超时和最大持续时间
- **传输监控**：设置最小传输速率和慢客户端检测

### 5. IP黑名单管理 (IPBlacklistSettings.js)

- **黑名单配置**：设置封禁时长和违规次数限制
- **IP管理**：手动添加/移除IP，查看封禁记录
- **封禁类型**：支持临时和永久封禁

### 6. 安全日志 (SecurityLogs.js)

- **日志查看**：分页显示安全事件日志
- **筛选功能**：按类型、IP、时间范围筛选
- **详情展示**：显示事件详细信息和处理结果

## 使用方法

### 1. 集成到设置页面

已经集成到 `web/src/pages/Setting/index.js` 中：

```javascript
import SecuritySetting from '../../components/SecuritySetting.js';

// 在导航菜单中添加
{ itemKey: 'security', text: t('安全设置'), icon: <IconShield /> }

// 在渲染函数中添加
case 'security':
  return <SecuritySetting />;
```

### 2. 访问安全设置

1. 以管理员身份登录 Tea API
2. 进入"设置"页面
3. 点击左侧导航的"安全设置"
4. 即可看到完整的安全配置界面

### 3. 配置安全防护

#### 异常检测配置
```javascript
{
  enabled: true,                    // 启用异常检测
  max_prompt_length: 50000,         // 最大Prompt长度
  max_random_char_ratio: 0.8,       // 随机字符比例阈值
  suspicious_score_limit: 100,      // 可疑分数限制
  auto_blacklist_enabled: true      // 自动加入黑名单
}
```

#### 请求限制配置
```javascript
{
  enabled: true,                    // 启用请求限制
  max_request_body_size: 10485760,  // 10MB
  max_prompt_length: 100000,        // 最大Prompt长度
  max_messages_count: 100,          // 最大消息数量
  content_validation: true          // 内容质量验证
}
```

## API接口

### 获取安全配置
```javascript
GET /api/security/config
```

### 更新安全配置
```javascript
PUT /api/security/config
Content-Type: application/json

{
  "abnormal_detection": { ... },
  "request_size_limit": { ... },
  "stream_protection": { ... },
  "ip_blacklist": { ... }
}
```

### 获取安全统计
```javascript
GET /api/security/stats
```

### IP黑名单管理
```javascript
// 添加IP到黑名单
POST /api/security/blacklist
{
  "ip": "192.168.1.100",
  "reason": "恶意攻击",
  "temporary": true
}

// 移除IP
DELETE /api/security/blacklist/192.168.1.100
```

## 样式和主题

组件使用 Semi Design 组件库，支持：

- **响应式布局**：适配不同屏幕尺寸
- **主题切换**：支持亮色/暗色主题
- **国际化**：支持中英文切换
- **无障碍访问**：符合WCAG标准

## 开发说明

### 依赖项

```json
{
  "@douyinfe/semi-ui": "^2.x",
  "@douyinfe/semi-icons": "^2.x",
  "react": "^18.x",
  "react-i18next": "^12.x"
}
```

### 本地开发

```bash
# 安装依赖
npm install

# 启动开发服务器
npm start

# 访问安全设置
http://localhost:3000/setting?tab=security
```

### 构建部署

```bash
# 构建生产版本
npm run build

# 部署到服务器
npm run deploy
```

## 演示页面

查看 `web/demo/security-demo.html` 了解完整的功能演示。

## 注意事项

1. **权限控制**：安全设置仅对管理员用户可见
2. **实时更新**：配置修改后会立即生效
3. **数据备份**：建议在修改配置前备份当前设置
4. **监控告警**：建议配置监控系统关注安全指标

## 故障排除

### 常见问题

1. **配置保存失败**
   - 检查用户权限
   - 验证配置格式
   - 查看网络连接

2. **数据加载失败**
   - 检查API服务状态
   - 验证认证token
   - 查看浏览器控制台错误

3. **界面显示异常**
   - 清除浏览器缓存
   - 检查CSS样式加载
   - 验证组件依赖

### 调试方法

```javascript
// 开启调试模式
localStorage.setItem('debug', 'security:*');

// 查看API请求
console.log('Security API calls:', window.securityAPI);

// 检查组件状态
console.log('Security config:', securityConfig);
```

## 更新日志

- **v1.0.0** - 初始版本，包含基础安全设置功能
- **v1.1.0** - 添加IP黑名单管理和安全日志
- **v1.2.0** - 增强异常检测和流保护功能

## 贡献指南

欢迎提交Issue和Pull Request来改进安全设置功能。

## 许可证

本项目采用 MIT 许可证。
