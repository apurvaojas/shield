package middleware

import (
	"log"
	"net/http"
	"org-forms-config-management/services"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
)

func JWTMiddleware(publicRoutes []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the request is public
		requestURI := c.Request.RequestURI
		// publicRoute.includes(requestURI)
		for _, route := range publicRoutes {
			log.Println("@@@@@@@@@@@@@@@@@@")
			log.Println(route)
			log.Println(c.FullPath())
			log.Println(requestURI)

			if route == c.FullPath() {
				c.Next()
				return
			}
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		tokenString := strings.Split(authHeader, " ")[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("userRole", claims["role"])
			c.Set("userName", claims["userName"])
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		context := services.GetUserContextInstance()
		context.SetUsername(c.GetString("userName"))

		c.Next()
	}
}
