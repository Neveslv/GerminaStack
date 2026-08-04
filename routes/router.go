package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func NewRouter(authHandler *handlers.AuthHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	api := router.Group("/api")
	api.POST("/login", authHandler.Login)
	api.POST("/login/2fa", authHandler.CompleteLogin)
	return router
}
