# SMTP中继服务

一个基于Go语言开发的SMTP中继服务，类似于smtp2go，解决家庭IP没有反向解析的问题。

## 特性

- 🚀 **完全容器化部署** - 基于Docker Compose的一键部署
- 🔐 **多密钥对认证** - 支持一个用户创建多个SMTP凭据
- 📊 **智能配额管理** - 支持日配额、小时配额、域名白名单等
- 🔄 **自动重试机制** - 智能识别临时错误并自动重试
- 📈 **实时监控** - 完整的邮件发送统计和状态监控
- 🛡️ **安全防护** - 多层频率限制、IP黑名单、认证保护
- 📚 **API文档** - 完整的Swagger API文档
- 🏗️ **微服务架构** - API、SMTP服务器、Worker分离部署

## 技术栈

- **后端**: Go + Gin框架
- **数据库**: MongoDB
- **缓存**: Redis
- **消息队列**: RabbitMQ
- **上游SMTP**: Postfix (容器化)
- **反向代理**: Nginx
- **容器化**: Docker + Docker Compose

## 架构图

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   客户端应用    │───▶│  SMTP中继服务   │───▶│  上游Postfix    │
│  (邮件发送)     │    │  (端口2525)     │    │  (端口25)       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   消息队列      │
                       │  (RabbitMQ)     │
                       └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │   邮件处理器    │
                       │   (Worker)      │
                       └─────────────────┘
```

## 快速开始

### 1. 克隆项目

```bash
git clone <repository-url>
cd smtp_replier
```

### 2. 配置环境变量

```bash
cp config.env.example config.env
# 编辑config.env文件，修改相关配置
```

### 3. 启动服务

```bash
# 构建并启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f
```

### 4. 验证部署

```bash
# 检查API服务
curl http://localhost:8080/health

# 检查SMTP服务
telnet localhost 2525

# 访问API文档
# 浏览器打开: http://localhost:8080/swagger/index.html
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| API服务 | 8080 | REST API接口 |
| SMTP中继 | 2525 | SMTP邮件接收 |
| SMTP提交 | 587 | SMTP提交端口 |
| SMTPS | 465 | SMTP SSL端口 |
| 上游Postfix | 25 | 实际邮件发送 |
| MongoDB | 27017 | 数据库 |
| Redis | 6379 | 缓存 |
| RabbitMQ | 5672 | 消息队列 |
| RabbitMQ管理 | 15672 | 队列管理界面 |
| Nginx | 80/443 | 反向代理 |

## 使用说明

### 1. 创建用户账户

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### 2. 创建SMTP凭据

```bash
# 先登录获取JWT Token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'

# 创建SMTP凭据
curl -X POST http://localhost:8080/api/smtp-credentials \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "我的第一个凭据",
    "daily_quota": 1000,
    "hourly_quota": 100
  }'
```

### 3. 使用SMTP发送邮件

```python
import smtplib
from email.mime.text import MIMEText

# 使用返回的SMTP凭据
smtp_username = "relay_12345678_abcd"  # 从API返回获取
smtp_password = "generated_password"    # 从API返回获取

# 连接SMTP服务器
server = smtplib.SMTP('your-server-ip', 2525)
server.starttls()
server.login(smtp_username, smtp_password)

# 发送邮件
msg = MIMEText('测试邮件内容')
msg['Subject'] = '测试邮件'
msg['From'] = 'sender@example.com'
msg['To'] = 'recipient@example.com'

server.send_message(msg)
server.quit()
```

## 管理功能

### 查看邮件统计

```bash
curl -X GET http://localhost:8080/api/mail-logs/stats \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

### 管理SMTP凭据

```bash
# 查看所有凭据
curl -X GET http://localhost:8080/api/smtp-credentials \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"

# 重置凭据密码
curl -X POST http://localhost:8080/api/smtp-credentials/{id}/reset-password \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## 配置说明

### 环境变量配置

主要配置项说明：

```env
# 数据库配置
MONGODB_URI=mongodb://mongodb:27017/smtp_relay

# 上游SMTP服务器（容器化Postfix）
UPSTREAM_SMTP_HOST=postfix
UPSTREAM_SMTP_PORT=25

# 配额限制
DAILY_QUOTA_DEFAULT=1000
HOURLY_QUOTA_DEFAULT=100

# 安全配置
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=1h
```

### Postfix配置

Postfix配置文件位于 `configs/postfix/` 目录：

- `main.cf` - 主配置文件
- `master.cf` - 服务配置文件

## 监控和日志

### 查看服务日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 查看特定服务日志
docker-compose logs -f smtp-relay-api
docker-compose logs -f smtp-relay-server
docker-compose logs -f smtp-relay-worker
```

### 监控队列状态

访问RabbitMQ管理界面：http://localhost:15672
- 用户名: guest
- 密码: guest

## 故障排除

### 常见问题

1. **SMTP连接失败**
   - 检查端口是否正确 (2525)
   - 确认防火墙设置
   - 验证SMTP凭据是否正确

2. **邮件发送失败**
   - 查看Worker日志
   - 检查上游Postfix状态
   - 确认配额是否用尽

3. **数据库连接失败**
   - 检查MongoDB容器状态
   - 验证连接字符串配置

### 重启服务

```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart smtp-relay-api
```

## 开发

### 本地开发环境

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 编译
go build -o bin/api cmd/api/main.go
go build -o bin/smtp cmd/smtp/main.go
go build -o bin/worker cmd/worker/main.go
```

### 代码结构

```
├── cmd/                 # 应用程序入口
│   ├── api/            # API服务
│   ├── smtp/           # SMTP服务器
│   └── worker/         # 邮件处理器
├── internal/           # 内部包
│   ├── api/            # API路由和处理器
│   ├── auth/           # 认证服务
│   ├── database/       # 数据库连接
│   ├── models/         # 数据模型
│   ├── queue/          # 消息队列
│   ├── smtp/           # SMTP服务器
│   └── worker/         # 邮件处理器
├── configs/            # 配置文件
├── scripts/            # 初始化脚本
└── docker-compose.yml  # 容器编排
```

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request！ 