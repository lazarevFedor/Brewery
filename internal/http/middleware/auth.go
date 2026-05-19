package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"Brewery/internal/http/handlers"
)

// AdminAuth — middleware, проверяет JWT токен в заголовке Authorization
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := handlers.ParseToken(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("admin", claims.Username)
		c.Next()
	}
}