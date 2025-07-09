package main

import (
	_ "shield/docs" // This line is needed for swagger

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitSwagger sets up swagger documentation routes
func InitSwagger(router *gin.Engine) {
	// Serve swagger docs at /swagger/*any
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}
