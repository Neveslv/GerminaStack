package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterAccountRoutes(api *gin.RouterGroup, account *handlers.AccountHandler, authenticated gin.HandlerFunc) {
	routes := api.Group("", authenticated)
	routes.GET("/me", account.GetProfile)
	routes.PATCH("/me", account.UpdateProfile)
	routes.GET("/me/preferences", account.GetPreferences)
	routes.PATCH("/me/preferences", account.UpdatePreferences)
}
