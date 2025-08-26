# 实时聊天系统（WebSocket + RabbitMQ + Redis）

基于 Go + WebSocket + RabbitMQ + Redis 实现的分布式实时聊天系统，支持单聊、群聊、用户在线状态管理，无硬编码数据，可直接容器化部署。


## 一、项目核心功能
- **实时通信**：基于 WebSocket 实现客户端与服务器的长连接，低延迟消息推送；
- **消息路由**：通过 RabbitMQ 实现跨节点消息分发（支持单聊 `once`、群聊 `group` 类型）；
- **状态管理**：Redis 维护用户 ID 复用、节点映射、在线用户统计，确保分布式场景下的状态一致性；
- **优雅关闭**：支持用户主动退出清理、主程序优雅关闭（清理 Redis 数据、关闭 MQ 连接、释放用户连接）；
- **容器化部署**：提供 Dockerfile，支持快速构建镜像，适配国内网络（国内镜像源配置）。


## 二、技术栈
| 模块         | 技术选型                | 作用                     |
|--------------|-------------------------|--------------------------|
| 后端框架     | Gin                     | HTTP 服务与路由管理      |
| 实时通信     | Gorilla WebSocket       | WebSocket 连接与消息处理 |
| 消息队列     | RabbitMQ                | 跨节点消息路由           |
| 缓存/状态存储 | Redis                   | 用户 ID 复用、在线状态管理 |
| 日志         | Zap                     | 结构化日志记录           |
| 容器化       | Docker                  | 环境一致性与快速部署     |


## 三、环境依赖
1. Go 1.25（项目 `go.mod` 已指定，需对应版本）；
2. Redis 6.0+（用于状态存储、ID 管理）；
3. RabbitMQ 3.10+（用于消息分发）；
4. Docker 20.0+（可选，用于容器化部署）。

## 四、项目结构
```
├── .idea/            # IDE 配置（已过滤，不上传仓库）
├── gin/              # 项目核心代码目录
│   ├── Dockerfile    # Docker 构建配置（国内镜像源适配）
│   ├── go.mod        # Go 模块依赖
│   ├── main.go       # 程序入口（初始化、服务启动）
│   ├── api/          # 接口层（路由、处理器、中间件）
│   │   ├── router.go # 路由注册（如 /chat_home、/ping）
│   │   ├── handler/  # 业务处理器（ChatHome.go 处理 WebSocket 连接）
│   │   ├── middleware/ # 中间件（如跨域处理 origin.go）
│   │   └── request/  # 请求模型（如 ChatHome 接口的请求结构体）
│   ├── conf/         # 配置层（Conf.go 读取环境变量/配置文件）
│   ├── global/       # 全局资源（模型、工具包）
│   │   ├── model/    # 数据模型（Message.go 消息结构、model.go 全局变量）
│   │   ├── pkg/      # 工具包（RabbitMq.go 封装 MQ 操作）
│   │   └── text/     # 文本常量（若后续有固定文本可放这里）
│   ├── inits/        # 初始化层（inits.go 初始化 Redis、MQ、日志等）
│   └── srv/          # 服务层（预留：若后续拆分业务逻辑可放这里）
├── .gitignore        # Git 过滤规则（关键：过滤冗余文件）
└── README.md         # 项目说明文档（本文件）
```
## 五、快速启动步骤
### 1. 本地启动（开发环境）
#### 步骤 1：初始化依赖
```bash
# 进入 gin 目录（项目核心代码目录）
cd gin
# 下载 Go 依赖（国内可配置 GOPROXY=https://goproxy.cn）
go mod download
#执行
go run ./main.go
```
### 1. 镜像启动（开发环境）
```bash
#必要参数
docker run /
-p [anypath]:8080/
-e RABBIT_MQ_URL="amqp://user_name:password@addr/database"/
-e REDIS_PASSWORD="your_redis_password"/
-e REDIS_ADDR="your_redis_addr"/
my-websocket:v1.4