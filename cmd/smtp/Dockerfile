# 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache git

# 复制go模块文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建SMTP服务
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smtp ./cmd/smtp

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1001 appgroup && \
    adduser -D -s /bin/sh -u 1001 -G appgroup appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/smtp .

# 创建SSL证书目录
RUN mkdir -p /etc/ssl/certs /etc/ssl/private && \
    chown -R appuser:appgroup /app /etc/ssl

# 切换到非root用户
USER appuser

# 暴露SMTP端口
EXPOSE 25 587 465

# 启动服务
CMD ["./smtp"] 