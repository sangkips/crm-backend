package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/internal/presentation/http/middleware"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	tenantService *service.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// GetCurrentTenant returns the current user's active tenant
func (h *TenantHandler) GetCurrentTenant(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == uuid.Nil {
		response.BadRequest(c, "No active tenant")
		return
	}

	tenant, err := h.tenantService.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Tenant retrieved successfully", gin.H{
		"tenant": tenant,
	})
}

// ListTenants returns all tenants for super admins, or only tenants the user belongs to
func (h *TenantHandler) ListTenants(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var tenants []entity.Tenant
	var err error

	// Super admins can see all tenants
	if IsSuperAdmin(c) {
		tenants, err = h.tenantService.ListAllTenants(c.Request.Context())
	} else {
		tenants, err = h.tenantService.GetUserTenants(c.Request.Context(), *userID)
	}

	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Tenants retrieved successfully", gin.H{
		"tenants": tenants,
	})
}

// UpdateTenant updates the current tenant's settings
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == uuid.Nil {
		response.BadRequest(c, "No active tenant")
		return
	}

	var req struct {
		Name     string                 `json:"name"`
		Settings *entity.TenantSettings `json:"settings"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	tenant, err := h.tenantService.UpdateTenant(c.Request.Context(), &service.UpdateTenantInput{
		ID:   tenantID,
		Name: req.Name,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Tenant updated successfully", gin.H{
		"tenant": tenant,
	})
}

// ListMembers returns all members of the current tenant
func (h *TenantHandler) ListMembers(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == uuid.Nil {
		response.BadRequest(c, "No active tenant")
		return
	}

	members, err := h.tenantService.GetTenantMembers(c.Request.Context(), tenantID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Members retrieved successfully", gin.H{
		"members": members,
	})
}

// InviteMember invites a user to the current tenant
func (h *TenantHandler) InviteMember(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == uuid.Nil {
		response.BadRequest(c, "No active tenant")
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Role   string    `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	err := h.tenantService.InviteMember(c.Request.Context(), &service.InviteMemberInput{
		TenantID: tenantID,
		UserID:   req.UserID,
		Role:     req.Role,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Member invited successfully", nil)
}

// RemoveMember removes a user from the current tenant
func (h *TenantHandler) RemoveMember(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == uuid.Nil {
		response.BadRequest(c, "No active tenant")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	if err := h.tenantService.RemoveMember(c.Request.Context(), tenantID, userID); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Member removed successfully", nil)
}

// UpdateMemberRole updates a member's role in the current tenant
func (h *TenantHandler) UpdateMemberRole(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == uuid.Nil {
		response.BadRequest(c, "No active tenant")
		return
	}

	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(c, "Invalid user ID")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.tenantService.UpdateMemberRole(c.Request.Context(), tenantID, userID, req.Role); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Member role updated successfully", nil)
}

// ListAllTenants returns all tenants (super admin only)
func (h *TenantHandler) ListAllTenants(c *gin.Context) {
	tenants, err := h.tenantService.ListAllTenants(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "All tenants retrieved successfully", gin.H{
		"tenants": tenants,
	})
}

// AssignUserToTenant assigns a user to a tenant (super admin only)
func (h *TenantHandler) AssignUserToTenant(c *gin.Context) {
	var req struct {
		TenantID uuid.UUID `json:"tenant_id" binding:"required"`
		UserID   uuid.UUID `json:"user_id" binding:"required"`
		Role     string    `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	// Default role to member if not specified
	if req.Role == "" {
		req.Role = "member"
	}

	err := h.tenantService.AssignUserToTenant(c.Request.Context(), &service.AssignUserToTenantInput{
		TenantID: req.TenantID,
		UserID:   req.UserID,
		Role:     req.Role,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "User assigned to tenant successfully", nil)
}
