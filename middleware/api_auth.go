package middleware

import (
	"net/http"
	"strconv"

	"germinaStack/auth"

	"github.com/gin-gonic/gin"
)

func APIAuthMiddleware(secret string) gin.HandlerFunc {
	return apiAuthMiddleware(secret, false)
}

func APIAdminAuthMiddleware(secret string) gin.HandlerFunc {
	return apiAuthMiddleware(secret, true)
}

func apiAuthMiddleware(secret string, requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(auth.CookieName)
		if err != nil || tokenString == "" {
			abortUnauthorized(c)
			return
		}

		claims, err := auth.ParseToken(tokenString, secret)
		if err != nil {
			abortUnauthorized(c)
			return
		}
		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil || userID <= 0 {
			abortUnauthorized(c)
			return
		}
		if requireAdmin && !claims.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "acesso proibido"})
			return
		}

		c.Set(auth.ContextUserID, userID)
		c.Set(auth.ContextIsAdmin, claims.IsAdmin)
		c.Next()
	}
}

func abortUnauthorized(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "n\u00e3o autorizado"})
}
