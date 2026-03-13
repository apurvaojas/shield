package swagger

import "github.com/gin-gonic/gin"

// ServeSwaggerJSON serves the generated swagger JSON file
// This endpoint allows external tools and clients to access the OpenAPI specification
func ServeSwaggerJSON(c *gin.Context) {
	c.File("./docs/swagger.json")
}

// RedirectToSwaggerUI redirects requests to /swagger to the swagger UI index page
// This provides a convenient shorthand for accessing the documentation UI
func RedirectToSwaggerUI(c *gin.Context) {
	c.Redirect(302, "/swagger/index.html")
}
