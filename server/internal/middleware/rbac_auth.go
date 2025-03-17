package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sunmoonstrand/go-react-blog/server/internal/service"
	"github.com/sunmoonstrand/go-react-blog/server/internal/utils/response"
	"go.uber.org/zap"
)

// RBACAuth RBAC权限控制中间件
func RBACAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户ID
		userID, exists := c.Get("user_id")
		if !exists {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 获取当前请求的路径和方法
		requestPath := c.Request.URL.Path
		requestMethod := c.Request.Method

		// 获取用户角色ID列表
		roleIDs, exists := c.Get("role_ids")
		if !exists {
			zap.L().Error("用户角色信息缺失", zap.Any("user_id", userID))
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		// 检查是否为超级管理员
		for _, roleID := range roleIDs.([]int) {
			// 角色ID为1表示超级管理员
			if roleID == 1 {
				c.Next()
				return
			}
		}

		// 检查用户是否有权限访问当前路径
		hasPermission, err := service.CheckPermission(userID.(int), roleIDs.([]int), requestPath, requestMethod)
		if err != nil {
			zap.L().Error("权限检查失败",
				zap.Any("user_id", userID),
				zap.Any("role_ids", roleIDs),
				zap.String("path", requestPath),
				zap.String("method", requestMethod),
				zap.Error(err),
			)
			response.ServerError(c, "权限检查失败")
			c.Abort()
			return
		}

		if !hasPermission {
			response.Forbidden(c, "权限不足")
			c.Abort()
			return
		}

		c.Next()
	}
}
