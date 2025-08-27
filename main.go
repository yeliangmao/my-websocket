package main

import (
	"Gin/api"
	"Gin/global/model"
	"Gin/inits"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// main 函数是程序的入口点
// main function is the entry point of the program
func main() {
	// 延迟执行Exit函数，确保程序退出前执行清理操作
	// Defer the execution of Exit function to ensure cleanup before program exits
	defer Exit()

	// 初始化程序配置和资源
	// Initialize program configuration and resources
	inits.Init()

	// 创建Gin默认路由引擎
	// Create Gin default router engine
	r := gin.Default()

	// 注册API路由
	// Register API routes
	api.Run(r)

	// 创建 HTTP Server 实例
	// Create HTTP Server instance
	srv := &http.Server{
		Addr:    ":8080", // 服务器监听地址
		Handler: r,       // 处理HTTP请求的处理器
	}

	// 在 goroutine 中启动服务器
	// Start the server in a goroutine
	go func() {
		// 启动HTTP服务器并监听端口
		// Start HTTP server and listen on port
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			model.Logger.Fatal("服务器启动失败", zap.Error(err))
		}
	}()
	// 等待中断信号以进行优雅关闭
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// SIGINT  (Ctrl+C)
	// SIGTERM (default signal sent by kill command)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	model.Logger.Info("正在关闭服务器...")
	// 设置一个 5 秒的超时上下文
	// Set a 5-second timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 使用 Shutdown 方法优雅关闭服务器
	// Gracefully shutdown the server using Shutdown method
	if err := srv.Shutdown(ctx); err != nil {
		// 需要zap：使用zap替代log.Fatal进行错误日志记录
		// Need zap: Use zap instead of log.Fatal for error logging
		model.Logger.Fatal("服务器强制关闭: ", zap.Error(err))
		// Need zap: Use zap instead of log.Fatal for error logging
		model.Logger.Fatal("Server forced to shutdown: ", zap.Error(err))
	}

	// 需要zap：使用zap替代log.Println进行日志记录
	// Need zap: Use zap instead of log.Println for logging
	model.Logger.Info("服务器已退出")
	// Need zap: Use zap instead of log.Println for logging
	model.Logger.Info("Server exited properly")
}

// Exit 函数用于程序退出前的清理操作
// Exit function is used for cleanup operations before program exits
func Exit() {
	// 关闭用户连接
	// Close user connections
	CloseUserConn()
	// 关闭RabbitMQ连接
	// Close RabbitMQ connections
	CloseRabbieMqConn()
	// 删除Redis中的数据
	// Delete data in Redis
	RedisDelete()
	model.Logger.Info("exit already")
}

// RedisDelete 用于删除Redis中的特定数据
// RedisDelete is used to delete specific data in Redis
func RedisDelete() {
	// 删除指定的Redis键
	// Delete specified Redis key
	model.RDB.Del(model.Ctx, model.OnlyMark)
	// 删除哈希表中的指定字段
	// Delete specified field in hash table
	model.RDB.HDel(model.Ctx, "Nodes", model.OnlyMark)
}

// CloseRabbieMqConn 用于关闭所有RabbitMQ连接
// CloseRabbieMqConn is used to close all RabbitMQ connections
func CloseRabbieMqConn() {
	// 遍历所有RabbitMQ连接并销毁
	// Iterate through all RabbitMQ connections and destroy them
	for s, mq := range model.RabbieMqPoll {
		mq.Destory()
		delete(model.RabbieMqPoll, s)
	}
}

// CloseUserConn 用于关闭所有用户连接
// CloseUserConn is used to close all user connections
func CloseUserConn() {
	// 向所有连接发送退出信号
	// Send exit signal to all connections
	for _, node := range model.ConnectionPool {
		node.ExitFlag.Do(
			func() {
				node.Exit <- true
			},
		)
	}
}
