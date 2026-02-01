package middleware

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
)

// ExtractTenantFromHost extracts tenant slug from subdomain
// e.g., "acme.investify.com" -> "acme"
func ExtractTenantFromHost(host string) (string, error) {
	// Remove port if present
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}

	parts := strings.Split(host, ".")
	if len(parts) < 3 {
		return "", errors.New("invalid subdomain")
	}
	return parts[0], nil
}

// TenantMiddleware validates tenant from subdomain and adds to context
func TenantMiddleware(tenantRepo repository.TenantRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract tenant from subdomain
		tenantSlug, err := ExtractTenantFromHost(c.Request.Host)
		if err != nil {
			// Allow requests without subdomain (for backwards compatibility during migration)
			// In production, you may want to enforce this
			c.Set("tenant_id", uuid.Nil)
			c.Next()
			return
		}

		// Lookup tenant by slug
		tenant, err := tenantRepo.GetBySlug(c.Request.Context(), tenantSlug)
		if err != nil || tenant == nil {
			response.NotFound(c, "Tenant not found")
			c.Abort()
			return
		}

		// Validate user has access to this tenant (if authenticated)
		userIDVal, exists := c.Get("user_id")
		if exists {
			userID, ok := userIDVal.(uuid.UUID)
			if ok && userID != uuid.Nil {
				isMember, _ := tenantRepo.IsMember(c.Request.Context(), tenant.ID, userID)
				if !isMember {
					response.Forbidden(c, "Access denied to this tenant")
					c.Abort()
					return
				}
			}
		}

		// Set tenant ID in Gin context (for middleware/handlers)
		c.Set("tenant_id", tenant.ID)
		c.Set("tenant", tenant)

		// Also set tenant ID in request context (for services/repositories)
		ctx := infraRepo.WithTenant(c.Request.Context(), tenant.ID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequireTenant ensures a valid tenant context exists
func RequireTenant() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get("tenant_id")
		if !exists {
			response.BadRequest(c, "Tenant context required")
			c.Abort()
			return
		}

		id, ok := tenantID.(uuid.UUID)
		if !ok || id == uuid.Nil {
			response.BadRequest(c, "Invalid tenant context")
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetTenantID retrieves the tenant ID from gin context
func GetTenantID(c *gin.Context) uuid.UUID {
	tenantID, exists := c.Get("tenant_id")
	if !exists {
		return uuid.Nil
	}
	id, ok := tenantID.(uuid.UUID)
	if !ok {
		return uuid.Nil
	}
	return id
}
