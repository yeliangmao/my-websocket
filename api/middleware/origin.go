package middleware

import "github.com/gin-gonic/gin"

// OriginMiddleware 跨域请求处理中间件
// OriginMiddleware handles CORS (Cross-Origin Resource Sharing) requests
func OriginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 允许的跨域源：此处保留"*"（允许所有源），若需限制可改为具体域名（如"https://your-domain.com"）
		// 1. Allowed origin: Keep "*" here (allows all origins); change to specific domain (e.g., "https://your-domain.com") if restriction is needed
		c.Header("Access-Control-Allow-Origin", "*")

		// 2. 仅保留业务必需的请求方法（移除了PUT、PATCH、DELETE，保留GET/OPTIONS；若需POST可自行添加）
		// 2. Keep only business-essential request methods (removed PUT, PATCH, DELETE; keep GET/OPTIONS; add POST manually if needed)
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		// 3. 关键：允许WebSocket握手必需的请求头（必须添加！否则握手失败）
		
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Upgrade, Connection, Sec-WebSocket-Key, Sec-WebSocket-Version")

		// 4. 允许前端读取WebSocket相关的响应头（可选，但建议添加）
		c.Header("Access-Control-Expose-Headers", "Sec-WebSocket-Accept")
		// 3. 处理预检请求（OPTIONS请求）：直接返回200状态码，避免预检失败
		// 3. Handle preflight request (OPTIONS request): Return 200 status code directly to avoid preflight failure
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		// 4. 继续执行后续中间件/处理器
		// 4. Proceed to the next middleware/handler
		c.Next()
	}
}
