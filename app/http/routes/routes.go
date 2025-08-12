package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/nishiki/backend-go/app/container"
	"github.com/nishiki/backend-go/app/http/controllers"
	"github.com/nishiki/backend-go/app/http/middleware"
)

func Setup(router *gin.Engine, appContainer *container.Container) {
	// Get dependencies from container
	logger := appContainer.GetLogger()
	authMiddleware := appContainer.GetAuthMiddleware()

	// CORS middleware
	router.Use(cors.New(cors.Config{
		//AllowOrigins:     []string{"http://localhost:3000", "https://localhost:3000"},
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Global middleware
	router.Use(middleware.LoggingMiddleware(logger))
	router.Use(middleware.ErrorHandlingMiddleware(logger))

	// Health check endpoint (no auth required)
	authController := controllers.NewAuthController(appContainer, logger)
	router.GET("/health", authController.HealthCheck)

	// Initialize controllers
	groupController := controllers.NewGroupController(appContainer, logger)
	userController := controllers.NewUserController(appContainer, logger)
	containerController := controllers.NewContainerController(appContainer, logger)
	collectionController := controllers.NewCollectionController(appContainer, logger)
	objectController := controllers.NewObjectController(appContainer, logger)

	// Auth routes (without auth middleware for OIDC endpoints)
	auth := router.Group("/auth")
	{
		// OIDC proxy endpoints (no auth required for frontend integration)
		auth.GET("/oidc-config", authController.GetOIDCConfig)
		auth.POST("/token", authController.ProxyTokenExchange)

		// Protected auth endpoints
		auth.GET("/me", authMiddleware.RequireAuth(), authController.GetCurrentUser)
	}

	// Group routes
	groups := router.Group("/groups")
	groups.Use(authMiddleware.RequireAuth())
	{
		groups.GET("", groupController.GetGroups)
		groups.POST("", groupController.CreateGroup)
		groups.GET("/:id", groupController.GetGroup)
		groups.GET("/:id/containers", groupController.GetGroupContainers)
		groups.GET("/:id/users", groupController.GetGroupUsers)
		groups.POST("/join", groupController.JoinGroup)
	}

	// User routes
	users := router.Group("/users")
	users.Use(authMiddleware.RequireAuth())
	{
		users.GET("/:id", userController.GetUser)
	}

	// Container routes
	containers := router.Group("/containers")
	containers.Use(authMiddleware.RequireAuth())
	{
		containers.GET("", containerController.GetContainers)
		containers.POST("", containerController.CreateContainer)
		containers.GET("/:id", containerController.GetContainer)
	}


	// Account routes (mapped to user functionality)
	accounts := router.Group("/accounts")
	accounts.Use(authMiddleware.RequireAuth())
	{
		accounts.GET("/:id", userController.GetUser)
		
		// Collections under accounts
		accounts.GET("/:id/collections", collectionController.GetCollections)
		accounts.POST("/:id/collections", collectionController.CreateCollection)
		accounts.GET("/:id/collections/:collection_id", collectionController.GetCollection)
		accounts.PUT("/:id/collections/:collection_id", collectionController.UpdateCollection)
		accounts.DELETE("/:id/collections/:collection_id", collectionController.DeleteCollection)
		
		// Collection objects
		accounts.GET("/:id/collections/:collection_id/objects", objectController.GetCollectionObjects)
		accounts.POST("/:id/collections/:collection_id/import", objectController.BulkImportToCollection)
		
		// Objects under accounts
		accounts.POST("/:id/objects", objectController.CreateObject)
		accounts.PUT("/:id/objects/:object_id", objectController.UpdateObject)
		accounts.DELETE("/:id/objects/:object_id", objectController.DeleteObject)
		
		// Bulk operations
		accounts.POST("/:id/import", objectController.BulkImport)
	}
}
