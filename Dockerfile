# ========================
# Stage 1: Builder
# ========================
FROM golang:1.26-alpine AS builder

# 设置工作目录
WORKDIR /app

# 拷贝 go.mod/go.sum 并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源代码
COPY . .

# 限制 CPU 并编译
ENV GOMAXPROCS=1
RUN go build -o app -p 1

# ========================
# Stage 2: Runtime
# ========================
FROM alpine:latest

# 工作目录
WORKDIR /app

# 复制编译好的二进制
COPY --from=builder /app/app .

# 开放端口
EXPOSE 8081

# 启动应用
CMD ["./app"]