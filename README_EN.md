# Real-Time Chat System (WebSocket + RabbitMQ + Redis)

A distributed real-time chat system implemented based on Go + WebSocket + RabbitMQ + Redis. It supports one-on-one chat, group chat, and user online status management. There is no hard-coded data, and it can be directly deployed in a containerized manner.

## I. Core Functions of the Project

- **Real-Time Communication**：Establishes long connections between clients and servers based on WebSocket to achieve low-latency message pushing;

- **Message Routing**：Implements cross-node message distribution through RabbitMQ (supports once type for one-on-one chat and group type for group chat);

- **Status Management**：Redis maintains user ID reuse, node mapping, and online user statistics to ensure status consistency in distributed scenarios;
- **Graceful Shutdown**：Supports active user logout cleanup and graceful shutdown of the main program (cleans up Redis data, closes MQ connections, and releases user connections);

- **Containerized Deployment**：Provides a Dockerfile to support rapid image building and is adapted to domestic networks (with domestic image source configuration).



## II. Technology Stack
| module         | Technical selection                | function                     |
|--------------|-------------------------|--------------------------|
| Back-end frame     | Gin                     | HTTP services and routing management      |
| Real-time communication     | Gorilla WebSocket       | WebSocket connection and message processing |
| Message queue     | RabbitMQ                | Cross-node message routing           |
| Cache/state store | Redis                   | User ID reuse, online presence management|
| log       | Zap                     | Structured logging           |
|Containerization       | Docker                  | Environment consistency and rapid deployment   |


## III. Environmental Dependencies
1. Go 1.25 (specified in the project's go.mod, the corresponding version is required);

2. Redis 6.0+ (used for status storage and ID management);

3. RabbitMQ 3.10+ (used for message distribution);

4.Docker 20.0+ (optional, used for containerized deployment).


## IV. Project Structure
```
├── .idea/            # IDE configuration (filtered, not uploaded to the repository)
├── gin/              # Project core code directory
│   ├── Dockerfile    # Docker build configuration (adapted to domestic mirror sources)
│   ├── go.mod        # Go module dependencies
│   ├── main.go       # Program entry point (initialization, service startup)
│   ├── api/          # API layer (routing, handlers, middleware)
│   │   ├── router.go # Route registration (e.g., /chat_home, /ping)
│   │   ├── handler/  # Business handlers (ChatHome.go handles WebSocket connections)
│   │   ├── middleware/ # Middleware (e.g., origin.go for CORS handling)
│   │   └── request/  # Request models (e.g., request structure for the ChatHome API)
│   ├── conf/         # Configuration layer (Conf.go reads environment variables/configuration files)
│   ├── global/       # Global resources (models, toolkits)
│   │   ├── model/    # Data models (Message.go for message structure, model.go for global variables)
│   │   ├── pkg/      # Toolkits (RabbitMq.go encapsulates MQ operations)
│   │   └── text/     # Text constants (for fixed text content if needed in subsequent development)
│   ├── inits/        # Initialization layer (inits.go initializes Redis, MQ, logs, etc.)
│   └── srv/          # Service layer (reserved: for business logic splitting in subsequent development)
├── .gitignore        # Git ignore rules (critical: filters redundant files)
└── README.md         # Project documentation (this file)
```
## V. Quick Start Steps

### 1. Local Startup (Development Environment)

```bash
# Enter the gin directory (the core code directory of the project)
cd gin
# Download Go dependencies (for domestic users, configure GOPROXY=https://goproxy.cn)
go mod download
# Run the program
go run ./main.go
```
### 2. Image Startup (Development Environment)
```bash
# Required parameters
docker run \
-p [anypath]:8080 \
-e RABBIT_MQ_URL="amqp://user_name:password@addr/database" \
-e REDIS_PASSWORD="your_redis_password" \
-e REDIS_ADDR="your_redis_addr" \
my-websocket:v1.4
```
# Note
In the "Image Startup" command, the backslashes (\) are used for line breaks to improve readability; ensure there is no extra space after the backslash when executing the command.
[anypath] in the port mapping parameter (-p) needs to be replaced with the actual local port you want to use (e.g., 8081:8080 means mapping local port 8081 to container port 8080).
Replace placeholders such as user_name, password, addr, database, your_redis_password, and your_redis_addr with actual configuration information of your RabbitMQ and Redis services.