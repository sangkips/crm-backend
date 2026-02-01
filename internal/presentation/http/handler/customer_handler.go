package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	customerService *service.CustomerService
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

// List handles listing customers (supports both page-based and cursor-based pagination)
func (h *CustomerHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	search := c.Query("search")
	isSuperAdmin := IsSuperAdmin(c)

	// Check if cursor-based pagination is requested
	if cursor := c.Query("cursor"); cursor != "" || c.Query("limit") != "" {
		h.listWithCursor(c, *userID, search, isSuperAdmin)
		return
	}

	// Default to page-based pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))

	params := &pagination.PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	// For super admins, skip tenant scope to see all customers
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

	result, err := h.customerService.ListCustomers(ctx, *userID, params, search, isSuperAdmin)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Customers retrieved successfully", result)
}

// listWithCursor handles listing customers with cursor-based pagination
func (h *CustomerHandler) listWithCursor(c *gin.Context, userID uuid.UUID, search string, isSuperAdmin bool) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	cursor := c.Query("cursor")
	direction := c.DefaultQuery("direction", "next")

	params := &pagination.CursorParams{
		Cursor:    cursor,
		Direction: pagination.CursorDirection(direction),
		Limit:     limit,
	}

	// For super admins, skip tenant scope to see all customers
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

	result, err := h.customerService.ListCustomersWithCursor(ctx, userID, params, search, isSuperAdmin)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, 200, "Customers retrieved successfully", result)
}

// Create handles creating a customer
func (h *CustomerHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Name          string  `json:"name" binding:"required"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		KRAPin        *string `json:"kra_pin"`
		Address       *string `json:"address"`
		AccountHolder *string `json:"account_holder"`
		AccountNumber *string `json:"account_number"`
		BankName      *string `json:"bank_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	customer, err := h.customerService.CreateCustomer(c.Request.Context(), &service.CreateCustomerInput{
		UserID:        *userID,
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		KRAPin:        req.KRAPin,
		Address:       req.Address,
		AccountHolder: req.AccountHolder,
		AccountNumber: req.AccountNumber,
		BankName:      req.BankName,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Customer created successfully", customer)
}

// Get handles getting a single customer
func (h *CustomerHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid customer ID")
		return
	}

	customer, err := h.customerService.GetCustomer(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Customer retrieved successfully", customer)
}

// Update handles updating a customer
func (h *CustomerHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid customer ID")
		return
	}

	var req struct {
		Name          *string `json:"name"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		KRAPin        *string `json:"kra_pin"`
		Address       *string `json:"address"`
		AccountHolder *string `json:"account_holder"`
		AccountNumber *string `json:"account_number"`
		BankName      *string `json:"bank_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	customer, err := h.customerService.UpdateCustomer(c.Request.Context(), &service.UpdateCustomerInput{
		UserID:        *userID,
		ID:            id,
		IsSuperAdmin:  isSuperAdmin,
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		KRAPin:        req.KRAPin,
		Address:       req.Address,
		AccountHolder: req.AccountHolder,
		AccountNumber: req.AccountNumber,
		BankName:      req.BankName,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Customer updated successfully", customer)
}

// Delete handles deleting a customer
func (h *CustomerHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid customer ID")
		return
	}

	if err := h.customerService.DeleteCustomer(c.Request.Context(), *userID, id, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// SupplierHandler handles supplier-related HTTP requests
type SupplierHandler struct {
	supplierService *service.SupplierService
}

// NewSupplierHandler creates a new supplier handler
func NewSupplierHandler(supplierService *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{supplierService: supplierService}
}

// List handles listing suppliers
func (h *SupplierHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))
	search := c.Query("search")

	params := &pagination.PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	// For super admins, skip tenant scope to see all suppliers
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

	result, err := h.supplierService.ListSuppliers(ctx, *userID, params, search, isSuperAdmin)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Suppliers retrieved successfully", result)
}

// Create handles creating a supplier
func (h *SupplierHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Name          string  `json:"name" binding:"required"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		Address       *string `json:"address"`
		ShopName      *string `json:"shopname"`
		KRAPin        *string `json:"kra_pin"`
		Type          string  `json:"type"`
		AccountHolder *string `json:"account_holder"`
		AccountNumber *string `json:"account_number"`
		BankName      *string `json:"bank_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	supplier, err := h.supplierService.CreateSupplier(c.Request.Context(), &service.CreateSupplierInput{
		UserID:        *userID,
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		ShopName:      req.ShopName,
		KRAPin:        req.KRAPin,
		Type:          req.Type,
		AccountHolder: req.AccountHolder,
		AccountNumber: req.AccountNumber,
		BankName:      req.BankName,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Supplier created successfully", supplier)
}

// Get handles getting a single supplier
func (h *SupplierHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid supplier ID")
		return
	}

	supplier, err := h.supplierService.GetSupplier(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Supplier retrieved successfully", supplier)
}

// Update handles updating a supplier
func (h *SupplierHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid supplier ID")
		return
	}

	var req struct {
		Name          *string `json:"name"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		Address       *string `json:"address"`
		ShopName      *string `json:"shopname"`
		KRAPin        *string `json:"kra_pin"`
		Type          *string `json:"type"`
		AccountHolder *string `json:"account_holder"`
		AccountNumber *string `json:"account_number"`
		BankName      *string `json:"bank_name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	supplier, err := h.supplierService.UpdateSupplier(c.Request.Context(), &service.UpdateSupplierInput{
		UserID:        *userID,
		ID:            id,
		IsSuperAdmin:  isSuperAdmin,
		Name:          req.Name,
		Email:         req.Email,
		Phone:         req.Phone,
		Address:       req.Address,
		ShopName:      req.ShopName,
		KRAPin:        req.KRAPin,
		Type:          req.Type,
		AccountHolder: req.AccountHolder,
		AccountNumber: req.AccountNumber,
		BankName:      req.BankName,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Supplier updated successfully", supplier)
}

// Delete handles deleting a supplier
func (h *SupplierHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid supplier ID")
		return
	}

	if err := h.supplierService.DeleteSupplier(c.Request.Context(), *userID, id, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
