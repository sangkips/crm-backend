package middleware

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
)

const (
	// IdempotencyKeyHeader is the HTTP header for idempotency keys
	IdempotencyKeyHeader = "Idempotency-Key"
	// IdempotencyKeyTTL is how long keys are valid
	IdempotencyKeyTTL = 24 * time.Hour
)

// IdempotencyConfig holds configuration for the idempotency middleware
type IdempotencyConfig struct {
	Repo repository.IdempotencyRepository
}

// responseWriter wraps gin.ResponseWriter to capture the response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Idempotency middleware prevents duplicate requests using idempotency keys
func Idempotency(config IdempotencyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to POST, PUT, PATCH methods
		if c.Request.Method != "POST" && c.Request.Method != "PUT" && c.Request.Method != "PATCH" {
			c.Next()
			return
		}

		// Get idempotency key from header
		idempotencyKey := c.GetHeader(IdempotencyKeyHeader)
		if idempotencyKey == "" {
			// No idempotency key provided, proceed normally
			c.Next()
			return
		}

		// Get user ID from context
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.Next()
			return
		}
		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			c.Next()
			return
		}

		// Check if this key was already processed
		existing, err := config.Repo.GetByKey(c.Request.Context(), idempotencyKey, userID)
		if err != nil {
			c.Next()
			return
		}

		// If key exists and not expired, return cached response
		if existing != nil && !existing.IsExpired() {
			c.Header("X-Idempotency-Replayed", "true")
			c.Data(existing.ResponseCode, "application/json", []byte(existing.ResponseBody))
			c.Abort()
			return
		}

		// Capture the response
		blw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process the request
		c.Next()

		// Store the idempotency key with response
		ikey := &entity.IdempotencyKey{
			Key:          idempotencyKey,
			UserID:       userID,
			Endpoint:     c.Request.Method + " " + c.FullPath(),
			ResponseCode: c.Writer.Status(),
			ResponseBody: blw.body.String(),
			ExpiresAt:    time.Now().Add(IdempotencyKeyTTL),
		}

		_ = config.Repo.Create(c.Request.Context(), ikey)
	}
}

// IdempotencyRequired is a stricter version that requires an idempotency key
func IdempotencyRequired(config IdempotencyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to POST methods
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		// Require idempotency key
		idempotencyKey := c.GetHeader(IdempotencyKeyHeader)
		if idempotencyKey == "" {
			c.JSON(400, gin.H{
				"success": false,
				"message": "Idempotency-Key header is required for this request",
			})
			c.Abort()
			return
		}

		// Get user ID from context
		userIDValue, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"success": false,
				"message": "User not authenticated",
			})
			c.Abort()
			return
		}
		userID, ok := userIDValue.(uuid.UUID)
		if !ok {
			c.JSON(401, gin.H{
				"success": false,
				"message": "Invalid user ID",
			})
			c.Abort()
			return
		}

		// Check if this key was already processed
		existing, err := config.Repo.GetByKey(c.Request.Context(), idempotencyKey, userID)
		if err != nil {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Failed to check idempotency key",
			})
			c.Abort()
			return
		}

		// If key exists and not expired, return cached response
		if existing != nil && !existing.IsExpired() {
			c.Header("X-Idempotency-Replayed", "true")

			// Parse and return the cached response
			var cachedResponse map[string]interface{}
			if err := json.Unmarshal([]byte(existing.ResponseBody), &cachedResponse); err == nil {
				c.JSON(existing.ResponseCode, cachedResponse)
			} else {
				c.Data(existing.ResponseCode, "application/json", []byte(existing.ResponseBody))
			}
			c.Abort()
			return
		}

		// Capture the response
		blw := &responseWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process the request
		c.Next()

		// Only store successful responses (2xx status codes)
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			ikey := &entity.IdempotencyKey{
				Key:          idempotencyKey,
				UserID:       userID,
				Endpoint:     c.Request.Method + " " + c.FullPath(),
				ResponseCode: c.Writer.Status(),
				ResponseBody: blw.body.String(),
				ExpiresAt:    time.Now().Add(IdempotencyKeyTTL),
			}

			_ = config.Repo.Create(c.Request.Context(), ikey)
		}
	}
}
