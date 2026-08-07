package routes

import (
	"germinaStack/handlers"
	"germinaStack/middleware"

	"github.com/gin-gonic/gin"
)

func NewRouter(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, catalog *handlers.CatalogHandler, discussion *handlers.DiscussionHandler, account *handlers.AccountHandler, jwtSecret string, adminHandlers ...*handlers.AdminHandler) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Static("/static", "./static")
	for route, file := range map[string]string{
		"/":             "index.html",
		"/admin":        "admin.html",
		"/cadastro":     "cadastro.html",
		"/login":        "login.html",
		"/materias":     "materias.html",
		"/perfil":       "perfil.html",
		"/post":         "post.html",
		"/preferencias": "preferencias.html",
		"/publicar":     "publicar.html",
	} {
		file := file
		router.GET(route, func(c *gin.Context) { c.File("./web/" + file) })
	}

	api := router.Group("/api")
	api.POST("/login", authHandler.Login)
	api.POST("/users", userHandler.Register)
	api.POST("/login/2fa", authHandler.CompleteLogin)
	api.POST("/logout", authHandler.Logout)
	RegisterCatalogRoutes(api, catalog, middleware.APIAuthMiddleware(jwtSecret))
	RegisterDiscussionRoutes(api, discussion, middleware.APIAuthMiddleware(jwtSecret))
	RegisterAccountRoutes(api, account, middleware.APIAuthMiddleware(jwtSecret))
	if len(adminHandlers) > 0 && adminHandlers[0] != nil {
		RegisterAdminRoutes(api, adminHandlers[0], middleware.APIAuthMiddleware(jwtSecret))
	}
	return router
}
