package conf

import (
	"Gin/global/model"
	"Gin/global/pkg"
	"os"

	"go.uber.org/zap"
)

// RedisAddr Redis服务器地址（格式：IP:端口，如127.0.0.1:6379）
// Redis server address (format: IP:Port, e.g. 127.0.0.1:6379)
var RedisAddr string

// RedisPassword Redis连接密码
// Password for Redis connection
var RedisPassword string

// RabbitMqUrl RabbitMQ连接URL（格式：amqp://用户名:密码@IP:端口/VirtualHost，如amqp://kuteng:kuteng@127.0.0.1:5672/kuteng）
// RabbitMQ connection URL (format: amqp://username:password@IP:Port/VirtualHost, e.g. amqp://kuteng:kuteng@127.0.0.1:5672/kuteng)
var RabbitMqUrl string

// ReadConf 配置读取入口函数，统一调用各组件配置读取方法
// ReadConf is the entry function for config reading, which uniformly calls config reading methods of each component
func ReadConf() {
	ReadRedisAddr()     // 读取Redis地址配置
	ReadRedisPassword() // 读取Redis密码配置
	ReadRabbitMqUrl()   // 读取RabbitMQ连接URL配置
}

// ReadRabbitMqUrl 读取RabbitMQ连接URL配置（从环境变量获取）
// ReadRabbitMqUrl reads RabbitMQ connection URL config (obtained from environment variable)
func ReadRabbitMqUrl() {
	// 从环境变量"RABBIT_MQ_URL"中获取RabbitMQ连接URL
	// Get RabbitMQ connection URL from environment variable "RABBIT_MQ_URL"
	RabbitMqUrl = os.Getenv("RABBIT_MQ_URL")

	// 若环境变量未配置，记录Fatal级日志并终止程序（核心配置缺失，服务无法启动）
	// If the environment variable is not configured, log Fatal-level error and terminate the program (core config missing, service cannot start)
	if RabbitMqUrl == "" {
		model.Logger.Fatal("Missing Config: RabbitMQ connection URL is not set", zap.String("Config Name", "RabbitMqUrl"), zap.String("Required Env Var", "RABBIT_MQ_URL"))
	}

	// 将读取到的RabbitMQ URL赋值给pkg包的全局MQURL变量，供RabbitMQ工具类使用
	// Assign the read RabbitMQ URL to the global MQURL variable in pkg package for RabbitMQ utility class
	pkg.MQURL = RabbitMqUrl
}

// ReadRedisPassword 读取Redis连接密码配置（从环境变量获取）
// ReadRedisPassword reads Redis connection password config (obtained from environment variable)
func ReadRedisPassword() {
	// 从环境变量"REDIS_PASSWORD"中获取Redis连接密码
	// Get Redis connection password from environment variable "REDIS_PASSWORD"
	RedisPassword = os.Getenv("REDIS_PASSWORD")

	// 若环境变量未配置，记录Fatal级日志并终止程序（Redis密码缺失，无法建立连接）
	// If the environment variable is not configured, log Fatal-level error and terminate the program (Redis password missing, cannot establish connection)
	if RedisPassword == "" {
		model.Logger.Fatal("Missing Config: Redis password is not set", zap.String("Config Name", "Redis Password"), zap.String("Required Env Var", "REDIS_PASSWORD"))
	}
}

// ReadRedisAddr 读取Redis服务器地址配置（从环境变量获取）
// ReadRedisAddr reads Redis server address config (obtained from environment variable)
func ReadRedisAddr() {
	// 从环境变量"REDIS_ADDR"中获取Redis服务器地址
	// Get Redis server address from environment variable "REDIS_ADDR"
	RedisAddr = os.Getenv("REDIS_ADDR")

	// 若环境变量未配置，记录Fatal级日志并终止程序（Redis地址缺失，无法建立连接）
	// If the environment variable is not configured, log Fatal-level error and terminate the program (Redis address missing, cannot establish connection)
	if RedisAddr == "" {
		model.Logger.Fatal("Missing Config: Redis server address is not set", zap.String("Config Name", "Redis Addr"), zap.String("Required Env Var", "REDIS_ADDR"))
	}
}
