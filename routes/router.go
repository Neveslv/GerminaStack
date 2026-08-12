package routes

import (
	"germinaStack/handlers"
	"germinaStack/middleware"

	"github.com/gin-gonic/gin"
)

type RouterDependencies struct {
	Catalog       *handlers.CatalogHandler
	Discussion    *handlers.DiscussionHandler
	Messages      *handlers.MessageHandler
	Reactions     *handlers.ReactionHandler
	Notifications *handlers.NotificationHandler
	Authenticated gin.HandlerFunc
}

func NewRouter(authHandler *handlers.AuthHandler, userHandler *handlers.UserHandler, args ...any) *gin.Engine {
	var catalog *handlers.CatalogHandler
	var discussion *handlers.DiscussionHandler
	var account *handlers.AccountHandler
	var jwtSecret string
	var adminHandler *handlers.AdminHandler
	var dependencies *RouterDependencies
	if len(args) == 1 {
		if value, ok := args[0].(RouterDependencies); ok {
			dependencies = &value
		}
		if value, ok := args[0].(*RouterDependencies); ok {
			dependencies = value
		}
	} else if len(args) >= 4 {
		catalog, _ = args[0].(*handlers.CatalogHandler)
		discussion, _ = args[1].(*handlers.DiscussionHandler)
		account, _ = args[2].(*handlers.AccountHandler)
		jwtSecret, _ = args[3].(string)
		if len(args) > 4 {
			adminHandler, _ = args[4].(*handlers.AdminHandler)
		}
	}
	if dependencies != nil {
		catalog, discussion = dependencies.Catalog, dependencies.Discussion
	}
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
	authenticated := middleware.APIAuthMiddleware(jwtSecret)
	if dependencies != nil && dependencies.Authenticated != nil {
		authenticated = dependencies.Authenticated
	}
	if catalog != nil {
		RegisterCatalogRoutes(api, catalog, authenticated)
	}
	if discussion != nil {
		RegisterDiscussionRoutes(api, discussion, authenticated)
	}
	if account != nil {
		RegisterAccountRoutes(api, account, authenticated)
	}
	if dependencies != nil {
		if dependencies.Messages != nil {
			RegisterMessageRoutes(api, dependencies.Messages, authenticated)
		}
		if dependencies.Reactions != nil {
			RegisterReactionRoutes(api, dependencies.Reactions, authenticated)
		}
		if dependencies.Notifications != nil {
			RegisterNotificationRoutes(api, dependencies.Notifications, authenticated)
		}
		RegisterUserAccountRoutes(api, userHandler, authenticated)
	}
	if adminHandler != nil {
		RegisterAdminRoutes(api, adminHandler, authenticated)
	}
	return router
}
