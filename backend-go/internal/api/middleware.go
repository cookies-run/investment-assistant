package api

import (
	"net/http"
	"stock-monitor/internal/service"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		claims, err := service.ParseToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func getUserID(c *gin.Context) uint {
	uid, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return uid.(uint)
}
