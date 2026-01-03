# 多阶段构建 - 构建阶段
FROM golang:1.21-alpine AS builder

# 设置工作目录
WORKDIR /build

# 安装必要的构建工具
RUN apk add --no-cache git gcc musl-dev

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 编译应用（静态链接，禁用CGO以获得纯静态二进制）
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o nav-admin .

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖和时区数据
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海
ENV TZ=Asia/Shanghai

# 创建非root用户
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/nav-admin .

# 创建必要的目录
RUN mkdir -p /app/data /app/uploads /app/static && \
    chown -R appuser:appuser /app

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV SERVER_PORT=8080 \
    SERVER_MODE=release \
    DB_PATH=/app/data/admin.db \
    UPLOAD_PATH=/app/uploads \
    NAV_JSON_PATH=/app/static/nav.json

# 启动应用
CMD ["./nav-admin"]
