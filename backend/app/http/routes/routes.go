package routes

import (
	"net/http"

	"github.com/nishiki/backend/app/container"
	"github.com/nishiki/backend/app/http/controllers"
	"github.com/nishiki/backend/app/http/httputil"
	"github.com/nishiki/backend/app/http/middleware"
	"github.com/nishiki/backend/app/http/openapi"
)

// Setup configures all routes and returns an http.Handler
func Setup(appContainer *container.Container) http.Handler {
	// Get dependencies from container
	logger := appContainer.GetLogger()
	authMiddleware := appContainer.GetAuthMiddleware()

	// Create mux
	mux := http.NewServeMux()

	// Initialize controllers
	authController := controllers.NewAuthController(appContainer, logger)
	groupController := controllers.NewGroupController(appContainer, logger)
	userController := controllers.NewUserController(appContainer, logger)
	containerController := controllers.NewContainerController(appContainer, logger)
	collectionController := controllers.NewCollectionController(appContainer, logger)
	objectController := controllers.NewObjectController(appContainer, logger)

	// Define global middleware chain
	globalMiddleware := httputil.Chain(
		middleware.CORSMiddleware(middleware.CORSConfig{
			AllowOrigins:     []string{"*"},
			AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
		}),
		middleware.RecoveryMiddleware(logger),
		middleware.LoggingMiddleware(logger),
	)

	// Auth middleware for protected routes
	authRequired := authMiddleware.RequireAuth()

	// Helper to wrap handlers with auth middleware
	withAuth := func(h http.HandlerFunc) http.HandlerFunc {
		return httputil.WrapHandler(http.HandlerFunc(h), authRequired)
	}

	// API spec (no auth required — docs UI served by frontend)
	mux.HandleFunc("GET /api/openapi.json", openapi.HandleOpenAPISpec)

	// Health check endpoint (no auth required)
	mux.HandleFunc("GET /health", authController.HealthCheck)

	// Auth routes (without auth middleware for OIDC endpoints)
	mux.HandleFunc("GET /auth/oidc-config", authController.GetOIDCConfig)
	mux.HandleFunc("POST /auth/token", authController.ProxyTokenExchange)
	mux.HandleFunc("GET /auth/me", withAuth(authController.GetCurrentUser))

	// Group routes (all require auth)
	mux.HandleFunc("GET /groups", withAuth(groupController.GetGroups))
	mux.HandleFunc("POST /groups", withAuth(groupController.CreateGroup))
	mux.HandleFunc("GET /groups/{id}", withAuth(groupController.GetGroup))
	mux.HandleFunc("PUT /groups/{id}", withAuth(groupController.UpdateGroup))
	mux.HandleFunc("DELETE /groups/{id}", withAuth(groupController.DeleteGroup))
	mux.HandleFunc("GET /groups/{id}/containers", withAuth(groupController.GetGroupContainers))
	mux.HandleFunc("GET /groups/{id}/users", withAuth(groupController.GetGroupUsers))
	mux.HandleFunc("POST /groups/{id}/users/{user_id}", withAuth(groupController.AddGroupMember))
	mux.HandleFunc("DELETE /groups/{id}/users/{user_id}", withAuth(groupController.RemoveGroupMember))
	mux.HandleFunc("POST /groups/join", withAuth(groupController.JoinGroup))

	// User routes (all require auth)
	mux.HandleFunc("GET /users/{id}", withAuth(userController.GetUser))

	// Container routes (all require auth)
	mux.HandleFunc("GET /containers", withAuth(containerController.GetContainers))
	mux.HandleFunc("POST /containers", withAuth(containerController.CreateContainer))
	mux.HandleFunc("GET /containers/{container_id}", withAuth(containerController.GetContainer))
	mux.HandleFunc("PUT /containers/{container_id}", withAuth(containerController.UpdateContainer))

	// Account routes (mapped to user functionality, all require auth)
	mux.HandleFunc("GET /accounts/{id}", withAuth(userController.GetUser))

	// Collections under accounts
	mux.HandleFunc("GET /accounts/{id}/collections", withAuth(collectionController.GetCollections))
	mux.HandleFunc("POST /accounts/{id}/collections", withAuth(collectionController.CreateCollection))
	mux.HandleFunc("GET /accounts/{id}/collections/{collection_id}", withAuth(collectionController.GetCollection))
	mux.HandleFunc("PUT /accounts/{id}/collections/{collection_id}", withAuth(collectionController.UpdateCollection))
	mux.HandleFunc("DELETE /accounts/{id}/collections/{collection_id}", withAuth(collectionController.DeleteCollection))
	mux.HandleFunc("PUT /accounts/{id}/collections/{collection_id}/schema", withAuth(collectionController.UpdatePropertySchema))
	mux.HandleFunc("GET /accounts/{id}/collections/{collection_id}/export", withAuth(collectionController.ExportCollection))

	// Containers under collections
	mux.HandleFunc("GET /accounts/{id}/collections/{collection_id}/containers", withAuth(containerController.GetContainers))
	mux.HandleFunc("POST /accounts/{id}/collections/{collection_id}/containers", withAuth(containerController.CreateContainer))
	mux.HandleFunc("GET /accounts/{id}/collections/{collection_id}/containers/{container_id}", withAuth(containerController.GetContainer))
	mux.HandleFunc("PUT /accounts/{id}/collections/{collection_id}/containers/{container_id}", withAuth(containerController.UpdateContainer))
	mux.HandleFunc("DELETE /accounts/{id}/collections/{collection_id}/containers/{container_id}", withAuth(containerController.DeleteContainer))

	// Container objects
	mux.HandleFunc("GET /accounts/{id}/collections/{collection_id}/containers/{container_id}/objects", withAuth(objectController.GetCollectionObjects))
	mux.HandleFunc("DELETE /accounts/{id}/collections/{collection_id}/containers/{container_id}/objects/{object_id}", withAuth(objectController.RemoveObjectFromContainer))

	// Collection objects
	mux.HandleFunc("GET /accounts/{id}/collections/{collection_id}/objects", withAuth(objectController.GetCollectionObjects))
	mux.HandleFunc("POST /accounts/{id}/collections/{collection_id}/import", withAuth(objectController.BulkImportToCollection))

	// Bulk import to a container (container_id in request body)
	mux.HandleFunc("POST /accounts/{id}/import", withAuth(objectController.BulkImport))

	// Objects under accounts
	mux.HandleFunc("POST /accounts/{id}/objects", withAuth(objectController.CreateObject))
	mux.HandleFunc("PUT /accounts/{id}/objects/{object_id}", withAuth(objectController.UpdateObject))
	mux.HandleFunc("DELETE /accounts/{id}/objects/{object_id}", withAuth(objectController.DeleteObject))

	// Apply global middleware
	return globalMiddleware(mux)
}
