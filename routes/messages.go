package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterMessageRoutes(api *gin.RouterGroup, messages *handlers.MessageHandler, authenticated gin.HandlerFunc) {
	api.GET("/posts", authenticated, messages.ListPosts)
	api.POST("/posts", authenticated, messages.CreatePost)
	api.GET("/posts/:id", authenticated, messages.GetPost)
	api.GET("/posts/:id/comments", authenticated, messages.ListComments)
	api.POST("/posts/:id/comments", authenticated, messages.CreateComment)
	api.GET("/comments/:id/replies", authenticated, messages.ListReplies)
	api.POST("/comments/:id/replies", authenticated, messages.CreateReply)
}
