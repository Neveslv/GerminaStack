package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterUserAccountRoutes(api *gin.RouterGroup, users *handlers.UserHandler, authenticated gin.HandlerFunc) {
	api.GET("/users/me", authenticated, users.GetCurrentUser)
	api.PUT("/users/me/preferences", authenticated, users.UpdateAccessibilityPreferences)
	api.GET("/preferences", authenticated, users.GetAccessibilityPreferences)
	api.PUT("/preferences", authenticated, users.UpdateAccessibilityPreferences)

}
