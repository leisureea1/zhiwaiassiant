# XISU Go Backend

西外校园服务 Go 后端 - 从 NestJS 迁移而来的高性能后端服务

## 特性

- 🚀 高性能 - 基于 Gin 框架，比 NestJS 更快的响应速度
- 📦 轻量级 - 单一二进制文件，无需 Node.js 运行时
- 🔐 安全认证 - JWT 令牌认证，支持访问令牌和刷新令牌
- 👥 用户管理 - 完整的用户注册、登录、密码管理
- 📧 邮件服务 - 邮箱验证码、密码重置
- 📁 文件上传 - 头像上传、附件管理
- 📢 公告系统 - 公告发布、置顶、已读标记
- 🎓 教务系统 - 课程表、成绩、考试、考勤查询
- 👨‍💼 管理后台 - 用户管理、系统日志、功能开关
- 🔄 Redis 缓存 - 验证码、会话缓存

## 技术栈

- **Web 框架**: Gin
- **ORM**: GORM
- **数据库**: MySQL
- **缓存**: Redis
- **JWT**: golang-jwt/jwt
- **密码加密**: bcrypt
- **配置管理**: godotenv

## 快速开始

### 环境要求

- Go 1.22+
- MySQL 5.7+
- Redis 6.0+

### 安装

1. 克隆项目
```bash
git clone <repository-url>
cd backend-go
```

2. 安装依赖
```bash
go mod download
```

3. 配置环境变量
```bash
cp .env.example .env
# 编辑 .env 文件，填写数据库、Redis、邮件等配置
```

4. 运行数据库迁移
```bash
# 使用 Prisma 或直接导入 SQL
mysql -u root -p xisu < database/current_schema.sql
```

5. 启动服务
```bash
go run cmd/server/main.go
```

服务将在 `http://localhost:3000` 启动

### 编译

```bash
# 编译当前平台
go build -o xisu-backend cmd/server/main.go

# 交叉编译 Linux
GOOS=linux GOARCH=amd64 go build -o xisu-backend-linux cmd/server/main.go

# 交叉编译 Windows
GOOS=windows GOARCH=amd64 go build -o xisu-backend.exe cmd/server/main.go
```

## 项目结构

```
backend-go/
├── cmd/
│   └── server/          # 主程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库模型
│   ├── http/
│   │   ├── handlers/    # HTTP 处理器
│   │   ├── middleware/  # 中间件
│   │   └── response/    # 响应封装
│   └── service/         # 业务逻辑服务
│       └── jwxt/        # 教务系统服务
├── static/              # 静态文件
├── uploads/             # 上传文件目录
├── .env                 # 环境变量配置
├── .env.example         # 环境变量示例
├── go.mod               # Go 模块定义
└── README.md            # 项目文档
```

## API 文档

### 认证相关

#### 发送邮箱验证码
```http
POST /api/v1/auth/send-code
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### 验证邮箱验证码
```http
POST /api/v1/auth/verify-code
Content-Type: application/json

{
  "email": "user@example.com",
  "code": "123456"
}
```

#### 用户注册
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "password": "password123",
  "email": "john@example.com",
  "studentId": "20210001",
  "xiwaiPassword": "jwxt_password"
}
```

#### 用户登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "studentId": "20210001",
  "password": "password123"
}
```

#### 刷新令牌
```http
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refreshToken": "your-refresh-token"
}
```

#### 修改密码
```http
POST /api/v1/auth/change-password
Authorization: Bearer <access-token>
Content-Type: application/json

{
  "oldPassword": "old_password",
  "newPassword": "new_password"
}
```

#### 忘记密码
```http
POST /api/v1/auth/forgot-password
Content-Type: application/json

{
  "email": "user@example.com"
}
```

#### 重置密码
```http
POST /api/v1/auth/reset-password
Content-Type: application/json

{
  "token": "reset-token",
  "newPassword": "new_password"
}
```

### 用户相关

#### 获取当前用户信息
```http
GET /api/v1/users/me
Authorization: Bearer <access-token>
```

#### 更新用户信息
```http
PUT /api/v1/users/:id
Authorization: Bearer <access-token>
Content-Type: application/json

{
  "realName": "张三",
  "nickname": "小张",
  "college": "英文学院",
  "major": "英语",
  "className": "英语2101"
}
```

#### 上传头像
```http
POST /api/v1/users/:id/avatar/upload
Authorization: Bearer <access-token>
Content-Type: multipart/form-data

file: <image-file>
```

### 公告相关

#### 获取公告列表
```http
GET /api/v1/announcements?page=1&pageSize=20&type=NORMAL
Authorization: Bearer <access-token>
```

#### 获取公告详情
```http
GET /api/v1/announcements/:id
Authorization: Bearer <access-token>
```

#### 标记公告已读
```http
POST /api/v1/announcements/:id/mark-viewed
Authorization: Bearer <access-token>
```

### 教务系统相关

#### 获取课程表
```http
GET /api/v1/jwxt/course?semesterId=209
Authorization: Bearer <access-token>
```

#### 获取成绩
```http
GET /api/v1/jwxt/grade?semesterId=209
Authorization: Bearer <access-token>
```

#### 获取考试安排
```http
GET /api/v1/jwxt/exam?semesterId=209
Authorization: Bearer <access-token>
```

### 管理员相关

#### 获取仪表盘统计
```http
GET /api/v1/admin/dashboard/stats
Authorization: Bearer <access-token>
```

#### 获取系统日志
```http
GET /api/v1/admin/system-logs?page=1&pageSize=20&level=ERROR
Authorization: Bearer <access-token>
```

#### 创建公告
```http
POST /api/v1/admin/announcements
Authorization: Bearer <access-token>
Content-Type: application/json

{
  "title": "公告标题",
  "content": "公告内容",
  "type": "IMPORTANT",
  "isPinned": true
}
```

## 环境变量说明

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| APP_NAME | 应用名称 | XISU Go Backend |
| APP_ENV | 运行环境 | development |
| APP_PORT | 服务端口 | 3000 |
| API_PREFIX | API 前缀 | /api/v1 |
| CORS_ORIGINS | 允许的跨域源 | http://localhost:5173,http://localhost:3000 |
| DATABASE_URL | 数据库连接字符串 | - |
| REDIS_ADDR | Redis 地址 | 127.0.0.1:6379 |
| REDIS_PASSWORD | Redis 密码 | - |
| REDIS_DB | Redis 数据库编号 | 0 |
| JWT_SECRET | JWT 密钥 | - |
| JWT_REFRESH_SECRET | JWT 刷新令牌密钥 | - |
| JWT_ACCESS_EXPIRES | 访问令牌过期时间 | 15m |
| JWT_REFRESH_EXPIRES | 刷新令牌过期时间 | 168h |
| MAIL_HOST | SMTP 服务器地址 | smtp.gmail.com |
| MAIL_PORT | SMTP 端口 | 587 |
| MAIL_USERNAME | 邮箱用户名 | - |
| MAIL_PASSWORD | 邮箱密码/应用密码 | - |
| MAIL_FROM | 发件人地址 | - |

## 邮件配置说明

### Gmail

1. 启用两步验证
2. 生成应用专用密码：https://myaccount.google.com/apppasswords
3. 使用应用密码作为 `MAIL_PASSWORD`

### QQ 邮箱

1. 开启 SMTP 服务
2. 获取授权码
3. 配置：
   - MAIL_HOST=smtp.qq.com
   - MAIL_PORT=587
   - MAIL_PASSWORD=授权码

### 163 邮箱

1. 开启 SMTP 服务
2. 获取授权码
3. 配置：
   - MAIL_HOST=smtp.163.com
   - MAIL_PORT=465
   - MAIL_PASSWORD=授权码

## 部署

### Docker 部署

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o xisu-backend cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/xisu-backend .
COPY --from=builder /app/static ./static
EXPOSE 3000
CMD ["./xisu-backend"]
```

### 宝塔面板部署

参考 [DEPLOYMENT_BAOTA.md](../DEPLOYMENT_BAOTA.md)

### Systemd 服务

```ini
[Unit]
Description=XISU Go Backend
After=network.target

[Service]
Type=simple
User=www
WorkingDirectory=/www/wwwroot/xisu-backend
ExecStart=/www/wwwroot/xisu-backend/xisu-backend
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

## 性能对比

| 指标 | NestJS | Go |
|------|--------|-----|
| 启动时间 | ~3s | ~0.1s |
| 内存占用 | ~150MB | ~30MB |
| 并发处理 | ~5000 req/s | ~20000 req/s |
| 响应时间 | ~50ms | ~10ms |

## 开发

### 运行测试
```bash
go test ./...
```

### 代码格式化
```bash
go fmt ./...
```

### 代码检查
```bash
go vet ./...
```

## 迁移说明

本项目从 NestJS 迁移而来，保持了 API 接口的兼容性。详细迁移说明请参考 [MIGRATION_SUMMARY.md](./MIGRATION_SUMMARY.md)

## 许可证

MIT

## 贡献

欢迎提交 Issue 和 Pull Request！
