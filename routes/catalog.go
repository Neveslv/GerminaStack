package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterCatalogRoutes(api *gin.RouterGroup, catalog *handlers.CatalogHandler, admin gin.HandlerFunc) {
	api.GET("/years", catalog.ListYears)
	api.GET("/subjects", catalog.ListSubjects)
	adminRoutes := api.Group("/admin", admin)
	adminRoutes.POST("/years", catalog.CreateYear)
	adminRoutes.PATCH("/years/:id", catalog.UpdateYear)
	adminRoutes.DELETE("/years/:id", catalog.DeleteYear)
	adminRoutes.POST("/subjects", catalog.CreateSubject)
	adminRoutes.PATCH("/subjects/:id", catalog.UpdateSubject)
	adminRoutes.DELETE("/subjects/:id", catalog.DeleteSubject)
}
