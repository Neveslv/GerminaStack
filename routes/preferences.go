package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterPreferenceRoutes exposes the legacy client-compatible preference
// aliases without changing the /users/me account route surface.
func RegisterPreferenceRoutes(api *gin.RouterGroup, users *handlers.UserHandler, authenticated gin.HandlerFunc) {
	api.GET("/preferences", authenticated, users.GetAccessibilityPreferences)
	api.PUT("/preferences", authenticated, users.UpdateAccessibilityPreferences)
}
