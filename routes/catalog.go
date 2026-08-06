package routes

import (
	"germinaStack/handlers"

	"github.com/gin-gonic/gin"
)

func RegisterCatalogRoutes(api *gin.RouterGroup, catalog *handlers.CatalogHandler, authMiddleware gin.HandlerFunc) {
	api.GET("/years", catalog.ListYears)
	api.GET("/subjects", authMiddleware, catalog.ListSubjects)
}
