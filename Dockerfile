# Dockerfile
FROM golang:1.26-alpine

WORKDIR /app

# 复制 go.mod / go.sum 提前缓存依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译 Go 程序
RUN go build -o app

# 暴露端口
EXPOSE 8081

# 容器启动命令
CMD ["./app"]