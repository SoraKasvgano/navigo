package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		session, err := c.Cookie("session")
		if err != nil || session == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			c.Abort()
			return
		}

		// 可以在这里验证session的有效性
		// 简单实现：只要有session cookie就认为已登录
		c.Next()
	}
}
