package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterDiscussionRoutes(api *gin.RouterGroup, discussion *handlers.DiscussionHandler, authenticated gin.HandlerFunc) {
	routes := api.Group("", authenticated)
	routes.GET("/posts", discussion.ListPosts)
	routes.POST("/posts", discussion.CreatePost)
	routes.GET("/posts/:id", discussion.GetPost)
	routes.PATCH("/posts/:id", discussion.UpdatePost)
	routes.DELETE("/posts/:id", discussion.DeletePost)
	routes.GET("/posts/:id/comments", discussion.ListComments)
	routes.POST("/posts/:id/comments", discussion.CreateComment)
	routes.GET("/comments/:id/replies", discussion.ListReplies)
	routes.POST("/comments/:id/replies", discussion.CreateReply)
	routes.PATCH("/comments/:id", discussion.UpdateComment)
	routes.DELETE("/comments/:id", discussion.DeleteComment)
	routes.PATCH("/replies/:id", discussion.UpdateReply)
	routes.DELETE("/replies/:id", discussion.DeleteReply)
	routes.PUT("/posts/:id/reaction", discussion.React("post"))
	routes.PUT("/comments/:id/reaction", discussion.React("comment"))
	routes.PUT("/replies/:id/reaction", discussion.React("comment_on_comment"))
	routes.GET("/notifications", discussion.ListNotifications)
	routes.PATCH("/notifications/read-all", discussion.MarkNotificationsRead)
}
