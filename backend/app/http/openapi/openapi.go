package openapi

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-openapi/runtime/middleware"
	swagno "github.com/go-swagno/swagno/v3"
	"github.com/go-swagno/swagno/v3/components/endpoint"
	"github.com/go-swagno/swagno/v3/components/http/response"
	"github.com/go-swagno/swagno/v3/components/parameter"
	"github.com/go-swagno/swagno/v3/components/security"
	"github.com/go-swagno/swagno/v3/components/tag"
	"github.com/nishiki/backend-go/app/http/request"
	httpresp "github.com/nishiki/backend-go/app/http/response"
)

var (
	openAPISpec     []byte
	openAPISpecOnce sync.Once
)

// GenerateOpenAPISpec creates and caches the OpenAPI 3.0 JSON specification.
func GenerateOpenAPISpec() []byte {
	openAPISpecOnce.Do(func() {
		sw := swagno.New(swagno.Config{
			Title:       "Nishiki Inventory API",
			Version:     "1.0.0",
			Description: "Inventory management REST API with integrated MCP (Model Context Protocol) server. See x-mcp-tools, x-mcp-resources, and x-mcp-prompts for AI assistant integration.",
		})

		sw.SetBearerAuth("JWT", "Bearer token obtained from Authentik OIDC. Required for all endpoints except /health, /auth/oidc-config, and /auth/token.")

		sw.AddTags(
			tag.New("auth", "Authentication and session management"),
			tag.New("groups", "Group management and collaboration"),
			tag.New("users", "User profile and account information"),
			tag.New("collections", "Inventory collection management"),
			tag.New("containers", "Container and storage management"),
			tag.New("objects", "Inventory object CRUD operations"),
			tag.New("import", "Bulk import of inventory items"),
		)

		registerAuthEndpoints(sw)
		registerGroupEndpoints(sw)
		registerUserEndpoints(sw)
		registerCollectionEndpoints(sw)
		registerContainerEndpoints(sw)
		registerObjectEndpoints(sw)
		registerImportEndpoints(sw)

		baseSpec, err := sw.ToJson()
		if err != nil {
			baseSpec = []byte(`{"error":"Failed to generate OpenAPI spec"}`)
			openAPISpec = baseSpec
			return
		}

		// Inject MCP x-extensions into the spec
		var specMap map[string]interface{}
		if err := json.Unmarshal(baseSpec, &specMap); err != nil {
			openAPISpec = baseSpec
			return
		}
		specMap["x-mcp-tools"] = mcpToolsDocs()
		specMap["x-mcp-resources"] = mcpResourcesDocs()
		specMap["x-mcp-prompts"] = mcpPromptsDocs()
		specMap["x-mcp-config"] = mcpConfigExample()

		enriched, err := json.MarshalIndent(specMap, "", "  ")
		if err != nil {
			openAPISpec = baseSpec
			return
		}
		openAPISpec = enriched
	})
	return openAPISpec
}

// HandleOpenAPISpec serves the OpenAPI JSON specification at /api/openapi.json.
func HandleOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := GenerateOpenAPISpec()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(spec) //nolint:errcheck
}

// CreateRedocHandler returns an http.Handler serving Redoc documentation at /docs.
func CreateRedocHandler() http.Handler {
	return middleware.Redoc(middleware.RedocOpts{
		SpecURL: "/api/openapi.json",
		Title:   "Nishiki API Documentation",
	}, nil)
}

// authSecurity returns the security requirement for JWT-protected endpoints.
func authSecurity() []map[security.SecuritySchemeName][]string {
	return []map[security.SecuritySchemeName][]string{
		{"JWT": {}},
	}
}

// ============================================
// AUTH ENDPOINTS
// ============================================

func registerAuthEndpoints(sw *swagno.OpenAPI) {
	sw.AddEndpoints([]*endpoint.EndPoint{
		endpoint.New(
			endpoint.GET,
			"/health",
			endpoint.WithTags("auth"),
			endpoint.WithSummary("Health check"),
			endpoint.WithDescription("Returns server health status. No authentication required."),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(map[string]string{}, "200", "Server is healthy"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/auth/oidc-config",
			endpoint.WithTags("auth"),
			endpoint.WithSummary("Get OIDC configuration"),
			endpoint.WithDescription("Returns the OIDC provider configuration needed for frontend OAuth2 PKCE flow. No authentication required."),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(map[string]string{}, "200", "OIDC configuration"),
			}),
		),
		endpoint.New(
			endpoint.POST,
			"/auth/token",
			endpoint.WithTags("auth"),
			endpoint.WithSummary("Exchange authorization code for token"),
			endpoint.WithDescription("Proxies OAuth2 authorization code exchange to Authentik. No authentication required."),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(map[string]string{}, "200", "Access token response"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid authorization code"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/auth/me",
			endpoint.WithTags("auth"),
			endpoint.WithSummary("Get current user"),
			endpoint.WithDescription("Returns the currently authenticated user and their JWT claims."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.AuthInfoResponse{}, "200", "Current user and claims"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "401", "Not authenticated"),
			}),
		),
	})
}

// ============================================
// GROUP ENDPOINTS
// ============================================

func registerGroupEndpoints(sw *swagno.OpenAPI) {
	sw.AddEndpoints([]*endpoint.EndPoint{
		endpoint.New(
			endpoint.GET,
			"/groups",
			endpoint.WithTags("groups"),
			endpoint.WithSummary("List groups"),
			endpoint.WithDescription("Returns all groups the current user belongs to."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New([]httpresp.GroupResponse{}, "200", "List of groups"),
			}),
		),
		endpoint.New(
			endpoint.POST,
			"/groups",
			endpoint.WithTags("groups"),
			endpoint.WithSummary("Create group"),
			endpoint.WithDescription("Creates a new group. The creating user becomes the first member."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithBody(request.CreateGroupRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.GroupResponse{}, "201", "Created group"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/groups/{id}",
			endpoint.WithTags("groups"),
			endpoint.WithSummary("Get group"),
			endpoint.WithDescription("Returns details for a specific group."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Group ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.GroupResponse{}, "200", "Group details"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Group not found"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/groups/{id}/users",
			endpoint.WithTags("groups"),
			endpoint.WithSummary("List group members"),
			endpoint.WithDescription("Returns all users in a group."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Group ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New([]httpresp.UserResponse{}, "200", "List of group members"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Group not found"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/groups/{id}/containers",
			endpoint.WithTags("groups"),
			endpoint.WithSummary("List group containers"),
			endpoint.WithDescription("Returns all containers shared with a group."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Group ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New([]httpresp.ContainerResponse{}, "200", "List of containers"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Group not found"),
			}),
		),
		endpoint.New(
			endpoint.POST,
			"/groups/join",
			endpoint.WithTags("groups"),
			endpoint.WithSummary("Join group via invitation"),
			endpoint.WithDescription("Joins a group using an invitation hash code."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithBody(request.JoinGroupRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.JoinGroupResponse{}, "200", "Joined group successfully"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid or expired invitation hash"),
			}),
		),
	})
}

// ============================================
// USER ENDPOINTS
// ============================================

func registerUserEndpoints(sw *swagno.OpenAPI) {
	sw.AddEndpoints([]*endpoint.EndPoint{
		endpoint.New(
			endpoint.GET,
			"/users/{id}",
			endpoint.WithTags("users"),
			endpoint.WithSummary("Get user"),
			endpoint.WithDescription("Returns user profile information."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("User ID (UUID)")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.UserResponse{}, "200", "User profile"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "User not found"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/accounts/{id}",
			endpoint.WithTags("users"),
			endpoint.WithSummary("Get account (alias for user)"),
			endpoint.WithDescription("Returns user profile. Alias for GET /users/{id} used in account-scoped routes."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("User/Account ID (UUID)")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.UserResponse{}, "200", "User profile"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "User not found"),
			}),
		),
	})
}

// ============================================
// COLLECTION ENDPOINTS
// ============================================

func registerCollectionEndpoints(sw *swagno.OpenAPI) {
	sw.AddEndpoints([]*endpoint.EndPoint{
		endpoint.New(
			endpoint.GET,
			"/accounts/{id}/collections",
			endpoint.WithTags("collections"),
			endpoint.WithSummary("List collections"),
			endpoint.WithDescription("Returns all collections owned by or shared with the user."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.CollectionSummaryListResponse{}, "200", "List of collections with summary info"),
			}),
		),
		endpoint.New(
			endpoint.POST,
			"/accounts/{id}/collections",
			endpoint.WithTags("collections"),
			endpoint.WithSummary("Create collection"),
			endpoint.WithDescription("Creates a new inventory collection. object_type must be one of: food, book, videogame, music, boardgame, general."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
			),
			endpoint.WithBody(request.CreateCollectionRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.CollectionResponse{}, "201", "Created collection"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request or object_type"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/accounts/{id}/collections/{collection_id}",
			endpoint.WithTags("collections"),
			endpoint.WithSummary("Get collection"),
			endpoint.WithDescription("Returns a collection with all its containers and objects."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.CollectionResponse{}, "200", "Collection with containers and objects"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Collection not found"),
			}),
		),
		endpoint.New(
			endpoint.PUT,
			"/accounts/{id}/collections/{collection_id}",
			endpoint.WithTags("collections"),
			endpoint.WithSummary("Update collection"),
			endpoint.WithDescription("Updates a collection's name, tags, or location."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
			),
			endpoint.WithBody(request.UpdateCollectionRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.CollectionResponse{}, "200", "Updated collection"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request"),
				response.New(ErrorResponse{}, "404", "Collection not found"),
			}),
		),
		endpoint.New(
			endpoint.DELETE,
			"/accounts/{id}/collections/{collection_id}",
			endpoint.WithTags("collections"),
			endpoint.WithSummary("Delete collection"),
			endpoint.WithDescription("Deletes a collection and all its containers and objects."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(EmptyResponse{}, "200", "Collection deleted"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Collection not found"),
			}),
		),
	})
}

// ============================================
// CONTAINER ENDPOINTS
// ============================================

func registerContainerEndpoints(sw *swagno.OpenAPI) {
	sw.AddEndpoints([]*endpoint.EndPoint{
		// Top-level container routes
		endpoint.New(
			endpoint.GET,
			"/containers",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("List all containers"),
			endpoint.WithDescription("Returns all containers accessible to the current user."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New([]httpresp.ContainerResponse{}, "200", "List of containers"),
			}),
		),
		endpoint.New(
			endpoint.POST,
			"/containers",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("Create container"),
			endpoint.WithDescription("Creates a new container. type must be one of: room, bookshelf, shelf, binder, cabinet, general."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithBody(request.CreateContainerRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.ContainerResponse{}, "201", "Created container"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request or container type"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/containers/{container_id}",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("Get container"),
			endpoint.WithDescription("Returns a specific container with its objects."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("container_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Container ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.ContainerResponse{}, "200", "Container with objects"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Container not found"),
			}),
		),
		endpoint.New(
			endpoint.PUT,
			"/containers/{container_id}",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("Update container"),
			endpoint.WithDescription("Updates a container's properties."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("container_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Container ID")),
			),
			endpoint.WithBody(request.UpdateContainerRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.ContainerResponse{}, "200", "Updated container"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request"),
				response.New(ErrorResponse{}, "404", "Container not found"),
			}),
		),

		// Collection-scoped container routes
		endpoint.New(
			endpoint.GET,
			"/accounts/{id}/collections/{collection_id}/containers",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("List collection containers"),
			endpoint.WithDescription("Returns all containers within a specific collection."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New([]httpresp.ContainerResponse{}, "200", "List of containers"),
			}),
		),
		endpoint.New(
			endpoint.POST,
			"/accounts/{id}/collections/{collection_id}/containers",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("Create container in collection"),
			endpoint.WithDescription("Creates a new container within a specific collection."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
			),
			endpoint.WithBody(request.CreateContainerRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.ContainerResponse{}, "201", "Created container"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/accounts/{id}/collections/{collection_id}/containers/{container_id}",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("Get container in collection"),
			endpoint.WithDescription("Returns a specific container within a collection."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
				parameter.StrParam("container_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Container ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.ContainerResponse{}, "200", "Container with objects"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Container not found"),
			}),
		),
		endpoint.New(
			endpoint.PUT,
			"/accounts/{id}/collections/{collection_id}/containers/{container_id}",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("Update container in collection"),
			endpoint.WithDescription("Updates a specific container within a collection."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
				parameter.StrParam("container_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Container ID")),
			),
			endpoint.WithBody(request.UpdateContainerRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.ContainerResponse{}, "200", "Updated container"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request"),
				response.New(ErrorResponse{}, "404", "Container not found"),
			}),
		),
		endpoint.New(
			endpoint.DELETE,
			"/accounts/{id}/collections/{collection_id}/containers/{container_id}",
			endpoint.WithTags("containers"),
			endpoint.WithSummary("Delete container in collection"),
			endpoint.WithDescription("Deletes a container and all its objects from the collection."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
				parameter.StrParam("container_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Container ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(EmptyResponse{}, "200", "Container deleted"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Container not found"),
			}),
		),
	})
}

// ============================================
// OBJECT ENDPOINTS
// ============================================

func registerObjectEndpoints(sw *swagno.OpenAPI) {
	sw.AddEndpoints([]*endpoint.EndPoint{
		endpoint.New(
			endpoint.GET,
			"/accounts/{id}/collections/{collection_id}/objects",
			endpoint.WithTags("objects"),
			endpoint.WithSummary("List collection objects"),
			endpoint.WithDescription("Returns all objects within a collection."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(OpenAPIObjectListResponse{}, "200", "List of objects"),
			}),
		),
		endpoint.New(
			endpoint.GET,
			"/accounts/{id}/collections/{collection_id}/containers/{container_id}/objects",
			endpoint.WithTags("objects"),
			endpoint.WithSummary("List container objects"),
			endpoint.WithDescription("Returns all objects within a specific container."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
				parameter.StrParam("container_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Container ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(OpenAPIObjectListResponse{}, "200", "List of objects"),
			}),
		),
		endpoint.New(
			endpoint.POST,
			"/accounts/{id}/objects",
			endpoint.WithTags("objects"),
			endpoint.WithSummary("Create object"),
			endpoint.WithDescription("Creates a new inventory object. object_type must be one of: food, book, videogame, music, boardgame, general. Properties is a free-form map of type-specific fields."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
			),
			endpoint.WithBody(OpenAPICreateObjectRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(OpenAPICreateObjectResponse{}, "201", "Created object"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request or object_type"),
			}),
		),
		endpoint.New(
			endpoint.PUT,
			"/accounts/{id}/objects/{object_id}",
			endpoint.WithTags("objects"),
			endpoint.WithSummary("Update object"),
			endpoint.WithDescription("Updates an inventory object. container_id is required to locate the object."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("object_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Object ID")),
			),
			endpoint.WithBody(OpenAPIUpdateObjectRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(OpenAPIUpdateObjectResponse{}, "200", "Updated object"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid request"),
				response.New(ErrorResponse{}, "404", "Object not found"),
			}),
		),
		endpoint.New(
			endpoint.DELETE,
			"/accounts/{id}/objects/{object_id}",
			endpoint.WithTags("objects"),
			endpoint.WithSummary("Delete object"),
			endpoint.WithDescription("Deletes an inventory object."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("object_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Object ID")),
			),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.DeleteObjectResponse{}, "200", "Object deleted"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "404", "Object not found"),
			}),
		),
	})
}

// ============================================
// IMPORT ENDPOINTS
// ============================================

func registerImportEndpoints(sw *swagno.OpenAPI) {
	sw.AddEndpoints([]*endpoint.EndPoint{
		endpoint.New(
			endpoint.POST,
			"/accounts/{id}/collections/{collection_id}/import",
			endpoint.WithTags("import"),
			endpoint.WithSummary("Bulk import objects to collection"),
			endpoint.WithDescription("Imports multiple objects into an existing collection. distribution_mode controls container assignment: 'automatic' (auto-distribute), 'manual' (each item specifies container), 'target' (all to target_container_id). data is an array of objects where keys match the collection's object type fields."),
			endpoint.WithSecurity(authSecurity()),
			endpoint.WithParams(
				parameter.StrParam("id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Account/User ID")),
				parameter.StrParam("collection_id", parameter.Path, parameter.WithRequired(), parameter.WithDescription("Collection ID")),
			),
			endpoint.WithBody(OpenAPIBulkImportCollectionRequest{}),
			endpoint.WithSuccessfulReturns([]response.Response{
				response.New(httpresp.BulkImportResponse{}, "200", "Import results with counts and any errors"),
			}),
			endpoint.WithErrors([]response.Response{
				response.New(ErrorResponse{}, "400", "Invalid format or data"),
			}),
		),
	})
}

// ============================================
// MCP X-EXTENSIONS
// ============================================

type mcpTool struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	InputFields map[string]string `json:"input_fields,omitempty"`
}

type mcpResource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Template    bool   `json:"template,omitempty"`
}

type mcpPrompt struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Arguments   map[string]string `json:"arguments,omitempty"`
}

func mcpToolsDocs() []mcpTool {
	return []mcpTool{
		{Name: "create_collection", Description: "Create a new inventory collection for a specific object type (food, books, games, etc.)", InputFields: map[string]string{"name": "required", "object_type": "required: food|book|videogame|music|boardgame|general", "location": "optional", "group_id": "optional", "tags": "optional"}},
		{Name: "update_collection", Description: "Update a collection's name, location, or tags", InputFields: map[string]string{"collection_id": "required", "name": "optional", "location": "optional", "tags": "optional"}},
		{Name: "delete_collection", Description: "Delete a collection and all its containers and objects", InputFields: map[string]string{"collection_id": "required"}},
		{Name: "create_container", Description: "Create a new container within a collection", InputFields: map[string]string{"collection_id": "required", "name": "required", "type": "optional: room|bookshelf|shelf|binder|cabinet|general", "parent_container_id": "optional", "location": "optional", "capacity": "optional"}},
		{Name: "update_container", Description: "Update a container's name, type, location, or capacity", InputFields: map[string]string{"container_id": "required", "name": "optional", "type": "optional", "location": "optional", "capacity": "optional"}},
		{Name: "delete_container", Description: "Delete a container and all its objects", InputFields: map[string]string{"container_id": "required"}},
		{Name: "create_object", Description: "Add a new object to a container", InputFields: map[string]string{"container_id": "required", "name": "required", "object_type": "required", "description": "optional", "quantity": "optional", "unit": "optional", "tags": "optional", "expires_at": "optional (RFC3339)"}},
		{Name: "update_object", Description: "Update an existing inventory object", InputFields: map[string]string{"object_id": "required", "container_id": "required", "name": "optional", "quantity": "optional", "tags": "optional", "expires_at": "optional"}},
		{Name: "delete_object", Description: "Delete an inventory object", InputFields: map[string]string{"object_id": "required", "container_id": "required"}},
		{Name: "create_group", Description: "Create a new sharing group", InputFields: map[string]string{"name": "required", "description": "optional"}},
		{Name: "join_group", Description: "Join a group using an invitation hash", InputFields: map[string]string{"invitation_hash": "required"}},
		{Name: "update_group", Description: "Update a group's name or description", InputFields: map[string]string{"group_id": "required", "name": "optional", "description": "optional"}},
		{Name: "delete_group", Description: "Delete a group", InputFields: map[string]string{"group_id": "required"}},
		{Name: "bulk_import", Description: "Import multiple objects into a collection at once from structured data", InputFields: map[string]string{"collection_id": "required", "data": "required: array of object maps", "format": "required: json|csv", "distribution_mode": "optional: automatic|manual|target", "target_container_id": "optional"}},
	}
}

func mcpResourcesDocs() []mcpResource {
	return []mcpResource{
		{URI: "nishiki://health", Name: "health", Description: "Server health status"},
		{URI: "nishiki://me", Name: "me", Description: "Current authenticated user"},
		{URI: "nishiki://groups", Name: "groups", Description: "Groups the current user belongs to"},
		{URI: "nishiki://collections", Name: "collections", Description: "All collections owned by or shared with the current user"},
		{URI: "nishiki://containers", Name: "containers", Description: "All containers accessible to the current user"},
		{URI: "nishiki://collections/{id}", Name: "collection", Description: "A specific collection with its containers", Template: true},
		{URI: "nishiki://collections/{id}/containers", Name: "collection-containers", Description: "Containers within a specific collection", Template: true},
		{URI: "nishiki://collections/{id}/objects", Name: "collection-objects", Description: "Objects within a specific collection", Template: true},
		{URI: "nishiki://containers/{id}", Name: "container", Description: "A specific container with its objects", Template: true},
		{URI: "nishiki://containers/{id}/objects", Name: "container-objects", Description: "Objects within a specific container", Template: true},
		{URI: "nishiki://groups/{id}", Name: "group", Description: "A specific group with its members", Template: true},
		{URI: "nishiki://groups/{id}/containers", Name: "group-containers", Description: "Containers shared with a specific group", Template: true},
	}
}

func mcpPromptsDocs() []mcpPrompt {
	return []mcpPrompt{
		{Name: "inventory_summary", Description: "Full inventory summary: all collections, container counts, and object totals"},
		{Name: "add_receipt", Description: "Parse receipt items and bulk import them into the appropriate collection", Arguments: map[string]string{"receipt_text": "required: text content of the receipt to parse and import"}},
		{Name: "find_item", Description: "Search for an item across all collections and containers", Arguments: map[string]string{"query": "required: item name or description to search for"}},
		{Name: "expiration_check", Description: "Scan all food collections for items expiring soon", Arguments: map[string]string{"days": "optional: number of days ahead to check (default: 30)"}},
		{Name: "reorganize", Description: "Analyze inventory layout and suggest reorganization for better utilization", Arguments: map[string]string{"collection_id": "optional: ID of the collection to analyze (leave empty for all collections)"}},
	}
}

func mcpConfigExample() map[string]interface{} {
	return map[string]interface{}{
		"description": "Add to your MCP client config (e.g. ~/.claude/claude_desktop_config.json). Get NISHIKI_TOKEN by authenticating at your Authentik instance.",
		"example": map[string]interface{}{
			"mcpServers": map[string]interface{}{
				"nishiki": map[string]interface{}{
					"command": "/path/to/nishiki",
					"args":    []string{"--mcp"},
					"env": map[string]string{
						"NISHIKI_TOKEN": "eyJ...",
					},
				},
			},
		},
	}
}
