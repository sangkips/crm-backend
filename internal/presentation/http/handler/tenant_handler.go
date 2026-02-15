package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/internal/presentation/http/middleware"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	tenantService *service.TenantService
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *service.TenantService) *TenantHandler {
	return &TenantHandler{tenantService: tenantService}
}

// Create handles creating a new tenant
func (h *TenantHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
		Slug string `json:"slug"` // Optional
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), &service.CreateTenantInput{
		Name:    req.Name,
		Slug:    req.Slug,
		OwnerID: *userID,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Tenant created successfully", gin.H{
		"tenant": tenant,
	})
}

// GetCurrentTenant returns the current user's active tenant
// If no tenant is set in context, it retrieves the user's first/default tenant
func (h *TenantHandler) GetCurrentTenant(c *gin.Context) {
	// First try to get tenant from middleware context (for subdomain-based routing)
	tenantID := middleware.GetTenantID(c)
	if tenantID != uuid.Nil {
		tenant, err := h.tenantService.GetTenant(c.Request.Context(), tenantID)
		if err != nil {
			response.Error(c, err)
			return
		}
		response.OK(c, "Tenant retrieved successfully", gin.H{
			"tenant": tenant,
		})
		return
	}

	// Fallback: Get user's tenants and return the first one (default)
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// Get only the first tenant
	result, err := h.tenantService.GetUserTenants(c.Request.Context(), *userID, &pagination.PaginationParams{Page: 1, PerPage: 1})
	if err != nil {
		response.Error(c, err)
		return
	}

	if len(result.Tenants) == 0 {
		response.NotFound(c, "No tenant found for user")
		return
	}

	// Return the first tenant as the current/default tenant
	response.OK(c, "Tenant retrieved successfully", gin.H{
		"tenant": result.Tenants[0],
	})
}

// ListTenants returns all tenants for super admins, or only tenants the user belongs to
func (h *TenantHandler) ListTenants(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var params pagination.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Invalid pagination parameters")
		return
	}

	// Set default per_page to 10 if not provided
	if params.PerPage == 0 {
		params.PerPage = 10
	}
	// Validate other params (e.g. Page default to 1)
	params.Validate()

	var result *service.ListTenantsOutput
	var err error

	// Super admins can see all tenants
	if IsSuperAdmin(c) {
		result, err = h.tenantService.ListAllTenants(c.Request.Context(), &params)
	} else {
		result, err = h.tenantService.GetUserTenants(c.Request.Context(), *userID, &params)
	}

	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Tenants retrieved successfully", gin.H{
		"tenants":    result.Tenants,
		"pagination": pagination.NewPagination(result.Page, result.PerPage, result.Total),
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
	var params pagination.PaginationParams
	if err := c.ShouldBindQuery(&params); err != nil {
		response.BadRequest(c, "Invalid pagination parameters")
		return
	}

	if params.PerPage == 0 {
		params.PerPage = 10
	}
	params.Validate()

	result, err := h.tenantService.ListAllTenants(c.Request.Context(), &params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "All tenants retrieved successfully", gin.H{
		"tenants":    result.Tenants,
		"pagination": pagination.NewPagination(result.Page, result.PerPage, result.Total),
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
