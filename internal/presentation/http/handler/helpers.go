package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetUserID extracts the user ID from the Gin context
func GetUserID(c *gin.Context) *uuid.UUID {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return nil
	}
	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		return nil
	}
	return &userID
}

// GetUserEmail extracts the user email from the Gin context
func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get("user_email")
	if !exists {
		return ""
	}
	return email.(string)
}

// GetUserRoles extracts the user roles from the Gin context
func GetUserRoles(c *gin.Context) []string {
	roles, exists := c.Get("user_roles")
	if !exists {
		return nil
	}
	return roles.([]string)
}

// GetUserPermissions extracts the user permissions from the Gin context
func GetUserPermissions(c *gin.Context) []string {
	permissions, exists := c.Get("user_permissions")
	if !exists {
		return nil
	}
	return permissions.([]string)
}

// IsSuperAdmin checks if the user has the super-admin role
func IsSuperAdmin(c *gin.Context) bool {
	roles := GetUserRoles(c)
	for _, role := range roles {
		if role == "super-admin" {
			return true
		}
	}
	return false
}
