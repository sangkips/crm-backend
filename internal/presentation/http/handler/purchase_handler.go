package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// PurchaseHandler handles purchase-related HTTP requests
type PurchaseHandler struct {
	purchaseService *service.PurchaseService
}

// NewPurchaseHandler creates a new purchase handler
func NewPurchaseHandler(purchaseService *service.PurchaseService) *PurchaseHandler {
	return &PurchaseHandler{purchaseService: purchaseService}
}

// List handles listing purchases
func (h *PurchaseHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))
	search := c.Query("search")
	statusStr := c.Query("status")

	params := &repository.PurchaseFilterParams{
		Pagination: &pagination.PaginationParams{
			Page:    page,
			PerPage: perPage,
		},
		Search:         search,
		SortBy:         c.Query("sort_by"),
		SortOrder:      c.Query("sort_order"),
		SkipUserFilter: isSuperAdmin,
	}

	if statusStr != "" {
		statusInt, err := strconv.Atoi(statusStr)
		if err == nil {
			status := enum.PurchaseStatus(statusInt)
			params.Status = &status
		}
	}

	if supplierIDStr := c.Query("supplier_id"); supplierIDStr != "" {
		if supplierID, err := uuid.Parse(supplierIDStr); err == nil {
			params.SupplierID = &supplierID
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			params.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			params.EndDate = &endDate
		}
	}

	// For super admins, skip tenant scope to see all purchases
	ctx := c.Request.Context()
	if isSuperAdmin {
		ctx = infraRepo.WithSkipTenantScope(ctx, true)
		// Allow super admin to filter by specific tenant if provided
		if tenantIDStr := c.Query("tenant_id"); tenantIDStr != "" {
			if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
				ctx = infraRepo.WithTenant(ctx, tenantID)
				ctx = infraRepo.WithSkipTenantScope(ctx, false)
			}
		}
	}

	result, err := h.purchaseService.ListPurchases(ctx, *userID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Purchases retrieved successfully", result)
}

// Create handles creating a purchase
func (h *PurchaseHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		SupplierID    *uuid.UUID `json:"supplier_id"`
		TaxPercentage float64    `json:"tax_percentage"`
		Items         []struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
			UnitCost  float64   `json:"unit_cost"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	items := make([]service.PurchaseItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.PurchaseItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
		}
	}

	purchase, err := h.purchaseService.CreatePurchase(c.Request.Context(), &service.CreatePurchaseInput{
		UserID:        *userID,
		SupplierID:    req.SupplierID,
		TaxPercentage: req.TaxPercentage,
		Items:         items,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Purchase created successfully", purchase)
}

// Get handles getting a single purchase
func (h *PurchaseHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid purchase ID")
		return
	}

	purchase, err := h.purchaseService.GetPurchase(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Purchase retrieved successfully", purchase)
}

// Approve handles approving a purchase
func (h *PurchaseHandler) Approve(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid purchase ID")
		return
	}

	if err := h.purchaseService.ApprovePurchase(c.Request.Context(), *userID, id, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Purchase approved successfully", nil)
}

// Delete handles deleting a purchase
func (h *PurchaseHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid purchase ID")
		return
	}

	if err := h.purchaseService.DeletePurchase(c.Request.Context(), *userID, id, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// GetPending handles getting pending purchases
func (h *PurchaseHandler) GetPending(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))

	params := &pagination.PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	result, err := h.purchaseService.GetPendingPurchases(c.Request.Context(), *userID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Pending purchases retrieved successfully", result)
}
