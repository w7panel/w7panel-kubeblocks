# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum* ./

# 下载依赖
RUN go mod download

# 复制源码
COPY main.go .

# 构建
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o gin-k8s-app .

# 运行阶段
FROM alpine:latest

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/gin-k8s-app .

# 暴露端口
EXPOSE 8080

# 运行应用
CMD ["./gin-k8s-app"]
