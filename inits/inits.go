package inits

import (
	"Gin/conf"
	"Gin/global/model"
	"Gin/global/pkg"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Init 初始化程序所需的各种组件和资源
// Init initializes various components and resources required by the program
func Init() {
	conf.ReadConf()   // 读取配置文件
	ZapBuild()        // 初始化Zap日志
	RedisConn()       // 连接Redis
	RedisMakeBucket() // 初始化Redis中的LoginBucket
	PullAndConnRabbieMq()
	go TimingSynchronization() // 启动定时同步协程
	RabbitMqSumerConn()        // 初始化RabbitMQ消费者连接
}

// RabbitMqSumerConn 初始化多个RabbitMQ消费者
// RabbitMqSumerConn initializes multiple RabbitMQ consumers
func RabbitMqSumerConn() {
	// 启动10个消费者协程
	// Start 10 consumer goroutines
	for i := 0; i < 10; i++ {
		go RabbitMqSumerRun()
	}
}

// RabbitMqSumerRun 运行RabbitMQ消费者逻辑
// RabbitMqSumerRun runs the RabbitMQ consumer logic
func RabbitMqSumerRun() {
	// 消费简单队列
	// Consume simple queue
	simple, err := model.RabbieMqPoll[model.OnlyMark].ConsumeSimple()
	if err != nil {
		// 记录消费队列错误
		// Log error when consuming queue
		model.Logger.Error("ConsumeSimple", zap.Error(err))
		return
	}

	// 循环处理消息
	// Loop to process messages
	for delivery := range simple {
		var Response model.Response
		// 解析消息体
		// Parse message body
		json.Unmarshal(delivery.Body, &Response)

		// 根据消息类型分发数据
		// Distribute data according to message type
		switch Response.Type {
		case "group":
			// 群发消息，发送给所有节点
			// Group message, send to all nodes
			for _, node := range model.ConnectionPool {
				node.Data <- delivery.Body
			}
		case "once":
			// 单发消息，发送给目标节点
			// One-time message, send to target node
			model.ConnectionPool[Response.Target].Data <- delivery.Body
		}
	}
}

// TimingSynchronization 定时同步节点和RabbitMQ连接
// TimingSynchronization periodically synchronizes nodes and RabbitMQ connections
func TimingSynchronization() {
	// 初始同步一次
	// Initial synchronization
	// 创建定时器，每60秒同步一次
	// Create ticker, synchronize every 60 seconds
	ticker := time.Ticker{C: time.Tick(time.Second * 60)}
	for _ = range ticker.C {
		PullAndConnRabbieMq()
	}
}

// PullAndConnRabbieMq 拉取节点列表并同步RabbitMQ连接
// PullAndConnRabbieMq pulls node list and synchronizes RabbitMQ connections
func PullAndConnRabbieMq() {
	// 从Redis获取所有节点标识
	// Get all node identifiers from Redis
	OrderOnlyMark := model.RDB.HGetAll(model.Ctx, "Nodes").Val()
	ConnectedNodes := make(map[string]bool)
	// 为每个节点建立RabbitMQ连接（如果不存在）
	// Establish RabbitMQ connection for each node (if not exists)
	for s, _ := range OrderOnlyMark {
		_, ok := model.RabbieMqPoll[s]
		if !ok {
			Conn := pkg.NewRabbitMQSimple(s)
			model.Logger.Info("RabbieMq Conn", zap.String("node", s))
			model.RabbieMqPoll[s] = Conn
		}
		ConnectedNodes[s] = true
	}
	model.Logger.Info("Update RabbieMq Conn Already")
	// 移除已不存在的节点连接
	// Remove connections for non-existent nodes
	for s, _ := range model.RabbieMqPoll {
		if !ConnectedNodes[s] {
			model.RabbieMqPoll[s].Destory()
			delete(model.RabbieMqPoll, s)
		}
	}
}

// NodeIntoGroup 将当前节点加入Redis中的节点分组
// NodeIntoGroup adds current node to node group in Redis
func NodeIntoGroup() {
	model.RDB.HSet(model.Ctx, "Nodes", model.OnlyMark, "ok")
	model.Logger.Info("Node into redis already") // 记录Redis连接成功日志
}

// RedisMakeBucket 初始化Redis中的LoginBucket，填充10000个元素
// RedisMakeBucket initializes LoginBucket in Redis with 10000 elements
func RedisMakeBucket() {
	ok := model.RDB.Exists(model.Ctx, "LoginBucket").Val()
	if ok == 1 {
		return
	}
	for i := 0; i < 100; i++ {
		if err := model.RDB.LPush(model.Ctx, "LoginBucket", i).Err(); err != nil {
			// 记录Redis操作错误
			// Log Redis operation error
			model.Logger.Error("RedisMakeBucket", zap.Error(err))
		}
	}
}

// ZapBuild 初始化Zap日志实例
// ZapBuild initializes Zap logger instance
func ZapBuild() {
	model.Logger = zap.NewExample()
	model.Logger.Info("Zap build success") // 记录Zap初始化成功日志
}

// RedisConn 建立与Redis的连接
// RedisConn establishes connection to Redis
func RedisConn() {
	// 创建Redis客户端
	// Create Redis client
	model.RDB = redis.NewClient(&redis.Options{
		Addr:            conf.RedisAddr,     // Redis地址
		Password:        conf.RedisPassword, // Redis密码
		PoolSize:        50,                 // 连接池大小
		MinIdleConns:    10,                 // 最小空闲连接数
		MaxRetries:      3,                  // 最大重试次数
		DialTimeout:     5 * time.Second,    // 拨号超时时间
		ReadTimeout:     3 * time.Second,    // 读取超时时间
		WriteTimeout:    3 * time.Second,    // 写入超时时间
		ConnMaxIdleTime: 5 * time.Minute,    // 连接最大空闲时间
		PoolTimeout:     4 * time.Second,    // 连接池超时时间
	})

	// 测试Redis连接
	// Test Redis connection
	if err := model.RDB.Ping(model.Ctx).Err(); err != nil {
		// 记录Redis连接错误并终止程序
		// Log Redis connection error and terminate program
		model.Logger.Fatal("Redis connect error", zap.Error(err))
	}
	model.Logger.Info("Redis connect success") // 记录Redis连接成功日志
	NodeIntoGroup()
}
