package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sangkips/investify-api/internal/config"
)

// CORSMiddleware creates a CORS middleware with the provided configuration
func CORSMiddleware(cfg *config.CORSConfig) gin.HandlerFunc {
	corsConfig := cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     cfg.AllowedMethods,
		AllowHeaders:     cfg.AllowedHeaders,
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// If no origins are configured, allow common development origins
	if len(corsConfig.AllowOrigins) == 0 {
		corsConfig.AllowOrigins = []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://127.0.0.1:3000",
		}
	}

	// If no methods are configured, use defaults
	if len(corsConfig.AllowMethods) == 0 {
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	}

	// If no headers are configured, use defaults
	if len(corsConfig.AllowHeaders) == 0 {
		corsConfig.AllowHeaders = []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Request-ID",
			"Origin",
			"Idempotency-Key",
		}
	} else {
		// Ensure Idempotency-Key is in the allowed headers
		hasIdempotencyKey := false
		for _, h := range corsConfig.AllowHeaders {
			if h == "Idempotency-Key" {
				hasIdempotencyKey = true
				break
			}
		}
		if !hasIdempotencyKey {
			corsConfig.AllowHeaders = append(corsConfig.AllowHeaders, "Idempotency-Key")
		}
	}

	return cors.New(corsConfig)

}
