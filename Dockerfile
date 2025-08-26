# 第一阶段：编译阶段（使用国内Go模块代理加速依赖下载）
FROM golang:1.25-alpine AS builder

# 关键：设置国内Go模块代理（七牛云/阿里云代理，解决github拉取慢问题）
ENV GOPROXY=https://goproxy.cn,direct \
    GO111MODULE=on

WORKDIR /app
COPY go.mod .
COPY go.sum .

# 安装git（拉取依赖需要），并下载依赖（此时会走GOPROXY，速度极快）
RUN apk add --no-cache git && go mod download

# 复制代码并编译
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s" -o /app/main .


# 第二阶段：运行阶段（使用国内alpine镜像源）
FROM alpine:3.19 AS mywebstocketgo

# 关键：替换alpine默认源为国内阿里云源（加速tzdata等包的安装）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要工具（时区数据）
RUN apk add --no-cache tzdata

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

WORKDIR /app
# 从构建阶段复制二进制文件
COPY --from=builder --chown=appuser:appgroup /app/main .

# 设置时区（避免日志时间混乱）
ENV TZ=Asia/Shanghai

USER appuser
EXPOSE 8080
CMD ["/app/main"]
