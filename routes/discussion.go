package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterDiscussionRoutes(api *gin.RouterGroup, discussion *handlers.DiscussionHandler, authenticated gin.HandlerFunc) {
	routes := api.Group("", authenticated)
	routes.GET("/posts", discussion.ListPosts)
	routes.GET("/posts/:id/comments", discussion.ListComments)
	routes.GET("/comments/:id/replies", discussion.ListReplies)
	routes.PUT("/posts/:id/reaction", discussion.React("post"))
	routes.PUT("/comments/:id/reaction", discussion.React("comment"))
	routes.PUT("/replies/:id/reaction", discussion.React("comment_on_comment"))
	routes.GET("/notifications", discussion.ListNotifications)
	routes.PATCH("/notifications/read-all", discussion.MarkNotificationsRead)
}
