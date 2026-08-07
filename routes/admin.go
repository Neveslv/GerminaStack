package routes

import (
	"germinaStack/handlers"
	"github.com/gin-gonic/gin"
)

func RegisterAdminRoutes(api *gin.RouterGroup, admin *handlers.AdminHandler, authenticated gin.HandlerFunc) {
	routes := api.Group("/admin", authenticated)
	routes.GET("/users", admin.ListUsers)
	routes.GET("/posts", admin.ListPosts)
	routes.PATCH("/users/:id/ban", admin.BanUser)
	routes.PATCH("/users/:id/admin", admin.SetAdmin)
	routes.DELETE("/posts/:id", admin.DeletePost)
}
