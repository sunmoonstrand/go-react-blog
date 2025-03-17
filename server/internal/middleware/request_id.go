package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求ID中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取请求ID
		requestID := c.GetHeader("X-Request-ID")

		// 如果请求头中没有请求ID，则生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// 设置请求ID到上下文
		c.Set("X-Request-ID", requestID)

		// 设置请求ID到响应头
		c.Header("X-Request-ID", requestID)

		c.Next()
	}
}
