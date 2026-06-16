package middleware

import (
	"net/http"
	"os"

	"germinaStack/auth"

	"github.com/gin-gonic/gin"
)

func AutentMiddleware() gin.HandlerFunc {
	return authMiddleware(false)
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return authMiddleware(true)
}

func authMiddleware(requireAdmin bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie(auth.CookieName)
		if err != nil || tokenString == "" {
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			c.Redirect(http.StatusFound, "/login?msg=JWT+nao+configurado")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(tokenString, secret)
		if err != nil {
			c.Redirect(http.StatusFound, "/login?msg=Token+invalido")
			c.Abort()
			return
		}

		if requireAdmin && !claims.IsAdmin {
			c.Redirect(http.StatusFound, "/dashboard?msg=Acesso+restrito+a+administradores&type=error")
			c.Abort()
			return
		}

		c.Set(auth.ContextUserID, claims.Subject)
		c.Set(auth.ContextIsAdmin, claims.IsAdmin)
		c.Next()
	}
}