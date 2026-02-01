package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/sangkips/investify-api/internal/application/service"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
)

// DashboardHandler handles dashboard-related HTTP requests
type DashboardHandler struct {
	dashboardService *service.DashboardService
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetStats handles getting dashboard statistics
func (h *DashboardHandler) GetStats(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	// For super admins, skip tenant scope to see all data across tenants
	ctx := c.Request.Context()
	if IsSuperAdmin(c) {
		ctx = infraRepo.WithSkipTenantScope(ctx, true)
	}

	stats, err := h.dashboardService.GetDashboardStats(ctx, *userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Dashboard stats retrieved successfully", stats)
}
