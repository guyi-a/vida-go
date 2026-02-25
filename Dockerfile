# 多阶段构建：第一阶段 - 构建Go应用
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的构建工具
RUN apk add --no-cache git

# 配置Go代理（加速依赖下载）
ENV GOPROXY=https://goproxy.cn,direct

# 复制go.mod和go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译Go应用
# CGO_ENABLED=0: 禁用CGO，生成静态链接的二进制文件
# GOOS=linux: 目标操作系统为Linux
# -ldflags="-s -w": 去除调试信息，减小二进制文件大小
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o vida-api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o vida-worker ./cmd/worker

# 多阶段构建：第二阶段 - 运行环境
FROM alpine:latest

# 安装运行时依赖
# ca-certificates: HTTPS证书
# ffmpeg: 视频处理
# tzdata: 时区数据
RUN apk --no-cache add ca-certificates ffmpeg tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/vida-api .
COPY --from=builder /app/vida-worker .

# 复制配置文件
COPY --from=builder /app/configs ./configs

# 创建必要的目录
RUN mkdir -p /app/logs && \
    chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8000

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8000/healthz || exit 1

# 启动命令
CMD ["./vida-api"]
