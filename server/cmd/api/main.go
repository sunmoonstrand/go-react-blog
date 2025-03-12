package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 设置为发布模式
	// gin.SetMode(gin.ReleaseMode)

	// 创建一个默认的路由引擎
	r := gin.Default()

	// 定义一个GET请求处理器
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// 启动HTTP服务，默认在0.0.0.0:8080启动服务
	r.Run(":8080")
}
