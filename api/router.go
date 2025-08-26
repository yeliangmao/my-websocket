package api

import (
	"Gin/api/handler"
	"Gin/api/middleware"

	"github.com/gin-gonic/gin"
)

// Run 注册API路由
// Run registers API routes
func Run(r *gin.Engine) {
	// 创建路由组，并应用跨域中间件
	// Create a route group and apply origin middleware
	Origin := r.Group("", middleware.OriginMiddleware())

	// 注册ping接口，用于健康检查
	// Register ping interface for health check
	Origin.GET("/ping", handler.Ping)

	// 注册聊天首页接口
	// Register chat home page interface
	Origin.GET("/chat_home", handler.ChatHome)
}
