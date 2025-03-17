package middleware

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sunmoonstrand/go-react-blog/server/internal/model"
	"github.com/sunmoonstrand/go-react-blog/server/internal/utils/response"
)

// JWTAuth JWT认证中间件
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取Authorization头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			response.Unauthorized(c, "无效的认证格式")
			c.Abort()
			return
		}

		// 解析JWT
		tokenString := parts[1]
		claims := &model.CustomClaims{}

		// 解析token
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			// 验证签名算法
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("无效的签名算法")
			}
			return []byte(secret), nil
		})

		// 处理解析错误
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				response.Unauthorized(c, "登录已过期，请重新登录")
			} else {
				response.Unauthorized(c, "无效的认证令牌")
			}
			c.Abort()
			return
		}

		// 验证token有效性
		if !token.Valid {
			response.Unauthorized(c, "无效的认证令牌")
			c.Abort()
			return
		}

		// 检查令牌是否过期
		if claims.ExpiresAt.Time.Before(time.Now()) {
			response.Unauthorized(c, "登录已过期，请重新登录")
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role_ids", claims.RoleIDs)

		c.Next()
	}
}
