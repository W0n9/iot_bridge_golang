# 构建阶段
FROM golang:1.23-alpine AS builder

LABEL TsungWing Wong=<TsungWing_Wong@outlook.com>

WORKDIR /build

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 运行阶段
FROM alpine:3.21

WORKDIR /app

# 复制二进制文件
COPY --from=builder /build/main .

RUN touch config.yaml

# 暴露端口
EXPOSE 9580

# 启动命令
CMD ["./main"]