package routes

import (
	"germinaStack/handlers"
	"github.com/gin-gonic/gin"
)

// RegisterUserAccountRoutes registers the authenticated account surface.
// NewRouter composes this feature during the final integration issue.
func RegisterUserAccountRoutes(api *gin.RouterGroup, users *handlers.UserHandler, authenticated gin.HandlerFunc) {
	api.GET("/users/me", authenticated, users.GetCurrentUser)
	api.PUT("/users/me/preferences", authenticated, users.UpdateAccessibilityPreferences)
}
