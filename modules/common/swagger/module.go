package swagger

import (
	"shield/docs" // Imported to allow runtime configuration of SwaggerInfo

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ConfigureApiBasePath configures the API base path in the Swagger documentation metadata
// This ensures the API base path matches the router configuration
func ConfigureApiBasePath(basePath string) {
	// Ensure swagger BasePath reflects the API prefix used by the router
	docs.SwaggerInfo.BasePath = basePath
}

// RegisterSwaggerRoutes registers all swagger-related routes with the provided router
// This centralizes all swagger documentation endpoints in one place
func RegisterSwaggerRoutes(router gin.IRouter) {
	// Serve the generated swagger JSON file (under /docs)
	// This keeps the docs accessible and compatible with swagger UI
	router.GET("/docs/swagger.json", ServeSwaggerJSON)

	// Redirect bare /swagger to the index page served by gin-swagger
	router.GET("/swagger", RedirectToSwaggerUI)

	// Serve swagger UI with dynamically configured URL pointing to local JSON
	// This allows swagger UI to load the documentation from /docs/swagger.json
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json")))
}
