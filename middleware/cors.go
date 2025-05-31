package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"os"
)

func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendBaseURL != "" {
		// Only allow the configured frontend domain when credentials are enabled
		config.AllowOrigins = []string{frontendBaseURL}
		config.AllowCredentials = true
	} else {
		// Fallback to allow all origins but without credentials
		config.AllowAllOrigins = true
		config.AllowCredentials = false
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"*"}
	return cors.New(config)
}
