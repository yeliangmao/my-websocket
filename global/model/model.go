package model

import (
	"Gin/api/request"
	"Gin/global/pkg"
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Ctx 全局上下文实例，用于Redis等操作的上下文传递
// Ctx is a global context instance used for context passing in operations like Redis
var Ctx = context.Background()

// RDB Redis客户端实例，全局共享的Redis连接
// RDB is a Redis client instance, a globally shared Redis connection
var RDB *redis.Client

// OnlyMark 节点唯一标识，通过UUID生成，用于区分不同的服务节点
// OnlyMark is a unique node identifier generated via UUID to distinguish different service nodes
var OnlyMark = uuid.NewString()

// Logger Zap日志实例，用于全局日志记录
// Logger is a Zap logger instance used for global logging
var Logger *zap.Logger

// ConnectionPool 用户连接池，存储当前节点上所有在线用户的WebSocket连接信息
// ConnectionPool is a user connection pool that stores WebSocket connection information of all online users on the current node
var ConnectionPool = make(map[int]request.Node)

// RabbieMqPoll RabbitMQ连接池，存储当前节点与各个RabbitMQ队列的连接
// RabbieMqPoll is a RabbitMQ connection pool that stores connections between the current node and various RabbitMQ queues
var RabbieMqPoll = make(map[string]*pkg.RabbitMQ)
var ActiveConnWG sync.WaitGroup
