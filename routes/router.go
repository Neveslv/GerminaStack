package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	Catalog       *handlers.CatalogHandler
	Messages      *handlers.MessageHandler
	Reactions     *handlers.ReactionHandler
	Notifications *handlers.NotificationHandler
	Authenticated gin.HandlerFunc
}

func NewRouter(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, dependencies ...RouterDependencies) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api")
	api.POST("/login", authHandler.Login)
	api.POST("/users", userHandler.Register)
	api.POST("/login/2fa", authHandler.CompleteLogin)
	api.POST("/logout", authHandler.Logout)

	if len(dependencies) > 0 {
		dependency := dependencies[0]
		if dependency.Catalog != nil && dependency.Authenticated != nil {
			RegisterCatalogRoutes(api, dependency.Catalog, dependency.Authenticated)
		}
		if dependency.Messages != nil && dependency.Authenticated != nil {
			RegisterMessageRoutes(api, dependency.Messages, dependency.Authenticated)
		}
		if dependency.Authenticated != nil {
			RegisterUserAccountRoutes(api, userHandler, dependency.Authenticated)
			if dependency.Reactions != nil {
				RegisterReactionRoutes(api, dependency.Reactions, dependency.Authenticated)
			}
			if dependency.Notifications != nil {
				RegisterNotificationRoutes(api, dependency.Notifications, dependency.Authenticated)
			}
		}
	}
	return router
}
