package handler

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// QuotationHandler handles quotation-related HTTP requests
type QuotationHandler struct {
	quotationService *service.QuotationService
}

// NewQuotationHandler creates a new quotation handler
func NewQuotationHandler(quotationService *service.QuotationService) *QuotationHandler {
	return &QuotationHandler{quotationService: quotationService}
}

// CreateQuotationRequest represents the create quotation request body
type CreateQuotationRequest struct {
	CustomerID         *string                `json:"customer_id"`
	Date               string                 `json:"date" binding:"required"`
	TaxPercentage      float64                `json:"tax_percentage"`
	DiscountPercentage float64                `json:"discount_percentage"`
	ShippingAmount     float64                `json:"shipping_amount"`
	Note               *string                `json:"note"`
	Status             int                    `json:"status"`
	Items              []QuotationItemRequest `json:"items" binding:"required,min=1"`
}

// QuotationItemRequest represents a line item in the request
type QuotationItemRequest struct {
	ProductID string  `json:"product_id" binding:"required"`
	Quantity  int     `json:"quantity" binding:"required,min=1"`
	UnitPrice float64 `json:"unit_price" binding:"required"`
}

// List handles listing quotations
// @Summary List Quotations
// @Description Get all quotations with pagination and filtering
// @Tags quotations
// @Security BearerAuth
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Items per page"
// @Param search query string false "Search term"
// @Param status query int false "Status filter"
// @Success 200 {object} response.APIResponse
// @Router /quotations [get]
func (h *QuotationHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	page := 1
	perPage := 15
	if p := c.Query("page"); p != "" {
		if parsed, err := parsePositiveInt(p); err == nil {
			page = parsed
		}
	}
	if pp := c.Query("per_page"); pp != "" {
		if parsed, err := parsePositiveInt(pp); err == nil {
			perPage = parsed
		}
	}

	search := c.Query("search")

	var status *enum.QuotationStatus
	if s := c.Query("status"); s != "" {
		if parsed, err := parseNonNegativeInt(s); err == nil {
			st := enum.QuotationStatus(parsed)
			status = &st
		}
	}

	result, err := h.quotationService.ListQuotations(c.Request.Context(), &service.ListQuotationsInput{
		UserID:       *userID,
		IsSuperAdmin: isSuperAdmin,
		Pagination: &pagination.PaginationParams{
			Page:    page,
			PerPage: perPage,
		},
		Search: search,
		Status: status,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Quotations retrieved successfully", result)
}

// Get handles getting a single quotation
// @Summary Get Quotation
// @Description Get a quotation by ID
// @Tags quotations
// @Security BearerAuth
// @Produce json
// @Param id path string true "Quotation ID"
// @Success 200 {object} response.APIResponse
// @Router /quotations/{id} [get]
func (h *QuotationHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quotation ID")
		return
	}

	quotation, err := h.quotationService.GetQuotation(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Quotation retrieved successfully", quotation)
}

// Create handles creating a quotation
// @Summary Create Quotation
// @Description Create a new quotation
// @Tags quotations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body CreateQuotationRequest true "Quotation data"
// @Success 201 {object} response.APIResponse
// @Router /quotations [post]
func (h *QuotationHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req CreateQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.BadRequest(c, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Parse customer ID if provided
	var customerID *uuid.UUID
	if req.CustomerID != nil && *req.CustomerID != "" {
		parsed, err := uuid.Parse(*req.CustomerID)
		if err != nil {
			response.BadRequest(c, "Invalid customer ID")
			return
		}
		customerID = &parsed
	}

	// Parse items
	items := make([]service.QuotationItemInput, len(req.Items))
	for i, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			response.BadRequest(c, "Invalid product ID")
			return
		}
		items[i] = service.QuotationItemInput{
			ProductID: productID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	quotation, err := h.quotationService.CreateQuotation(c.Request.Context(), &service.CreateQuotationInput{
		UserID:             *userID,
		CustomerID:         customerID,
		Date:               date,
		TaxPercentage:      req.TaxPercentage,
		DiscountPercentage: req.DiscountPercentage,
		ShippingAmount:     req.ShippingAmount,
		Note:               req.Note,
		Status:             enum.QuotationStatus(req.Status),
		Items:              items,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Quotation created successfully", quotation)
}

// Update handles updating a quotation
// @Summary Update Quotation
// @Description Update an existing quotation
// @Tags quotations
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Quotation ID"
// @Param request body CreateQuotationRequest true "Quotation data"
// @Success 200 {object} response.APIResponse
// @Router /quotations/{id} [put]
func (h *QuotationHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quotation ID")
		return
	}

	var req CreateQuotationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		response.BadRequest(c, "Invalid date format. Use YYYY-MM-DD")
		return
	}

	// Parse customer ID if provided
	var customerID *uuid.UUID
	if req.CustomerID != nil && *req.CustomerID != "" {
		parsed, err := uuid.Parse(*req.CustomerID)
		if err != nil {
			response.BadRequest(c, "Invalid customer ID")
			return
		}
		customerID = &parsed
	}

	// Parse items
	items := make([]service.QuotationItemInput, len(req.Items))
	for i, item := range req.Items {
		productID, err := uuid.Parse(item.ProductID)
		if err != nil {
			response.BadRequest(c, "Invalid product ID")
			return
		}
		items[i] = service.QuotationItemInput{
			ProductID: productID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		}
	}

	quotation, err := h.quotationService.UpdateQuotation(c.Request.Context(), &service.UpdateQuotationInput{
		UserID:             *userID,
		ID:                 id,
		IsSuperAdmin:       isSuperAdmin,
		CustomerID:         customerID,
		Date:               date,
		TaxPercentage:      req.TaxPercentage,
		DiscountPercentage: req.DiscountPercentage,
		ShippingAmount:     req.ShippingAmount,
		Note:               req.Note,
		Status:             enum.QuotationStatus(req.Status),
		Items:              items,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Quotation updated successfully", quotation)
}

// Delete handles deleting a quotation
// @Summary Delete Quotation
// @Description Delete a quotation by ID
// @Tags quotations
// @Security BearerAuth
// @Param id path string true "Quotation ID"
// @Success 204
// @Router /quotations/{id} [delete]
func (h *QuotationHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid quotation ID")
		return
	}

	if err := h.quotationService.DeleteQuotation(c.Request.Context(), *userID, id, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// Helper functions for parsing query parameters
func parsePositiveInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil || result < 1 {
		return 1, err
	}
	return result, nil
}

func parseNonNegativeInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil || result < 0 {
		return 0, err
	}
	return result, nil
}
