package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterReactionRoutes(api *gin.RouterGroup, reactions *handlers.ReactionHandler, authenticated gin.HandlerFunc) {
	api.POST("/reactions", authenticated, reactions.Toggle)
	api.GET("/reactions", authenticated, reactions.Get)
}

func RegisterNotificationRoutes(api *gin.RouterGroup, notifications *handlers.NotificationHandler, authenticated gin.HandlerFunc) {
	api.GET("/notifications", authenticated, notifications.List)
	api.POST("/notifications/read", authenticated, notifications.MarkRead)
}
