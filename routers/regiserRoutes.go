package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	swaggerfiles "github.com/swaggo/files"

	ginSwagger "github.com/swaggo/gin-swagger"
)

// List of public routes
var PublicRoutes = []string{
	"/health",
	"/swagger/*any",
	"/metrics",
	"/api/v1/onboarding/signup",
	"/api/v1/onboarding/verifyEmail",
	"/api/v1/onboarding/resendConfirmationCode",
}

// RegisterRoutes add all routing list here automatically get main router
func RegisterRoutes(route *gin.Engine) {

	route.NoRoute(func(ctx *gin.Context) {
		ctx.JSON(http.StatusNotFound, gin.H{"status": http.StatusNotFound, "message": "Route Not Found"})
	})
	// docs.SwaggerInfo.BasePath = "/api/v1"
	route.GET("/health", func(ctx *gin.Context) { ctx.JSON(http.StatusOK, gin.H{"live": "ok"}) })
	route.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	route.GET("/metrics", gin.WrapH(promhttp.Handler()))
	//Add All route
	// ExamplesRoutes(route)
	CustomerOnboardingRoutes(route)
}
