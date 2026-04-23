# 多阶段构建

# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 设置Go模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 复制go mod和sum文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建沙盒服务器（包含完整功能）
RUN CGO_ENABLED=0 GOOS=linux go build -o risk-engine ./cmd/server/sandbox_server.go

# 生产阶段
FROM alpine:latest

# 安装ca证书（HTTPS请求需要）
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/risk-engine .

# 复制配置文件
COPY --from=builder /app/test/cases/simple/01-age-rule/config/ ./config/

# 确保sandbox records目录存在
RUN mkdir -p ./sandbox/records

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动服务
CMD ["./risk-engine"]
