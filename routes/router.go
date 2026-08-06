package routes

import (
	"germinaStack/handlers"
	"germinaStack/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, catalog *handlers.CatalogHandler, discussion *handlers.DiscussionHandler, account *handlers.AccountHandler, jwtSecret string) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api")
	api.POST("/login", authHandler.Login)
	api.POST("/users", userHandler.Register)
	api.POST("/login/2fa", authHandler.CompleteLogin)
	api.POST("/logout", authHandler.Logout)
	RegisterCatalogRoutes(api, catalog, middleware.APIAdminAuthMiddleware(jwtSecret))
	RegisterDiscussionRoutes(api, discussion, middleware.APIAuthMiddleware(jwtSecret))
	RegisterAccountRoutes(api, account, middleware.APIAuthMiddleware(jwtSecret))
	return router
}
