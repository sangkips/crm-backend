package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/request"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// ProductHandler handles product-related HTTP requests
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler creates a new product handler
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// List handles listing products (supports both page-based and cursor-based pagination)
func (h *ProductHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	// Check if cursor-based pagination is requested
	if cursor := c.Query("cursor"); cursor != "" || c.Query("limit") != "" {
		h.listWithCursor(c, *userID, isSuperAdmin)
		return
	}

	var filter request.ProductFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	params := &repository.ProductFilterParams{
		Pagination: &pagination.PaginationParams{
			Page:    filter.Page,
			PerPage: filter.PerPage,
		},
		Search:         filter.Search,
		LowStock:       filter.LowStock,
		SortBy:         filter.SortBy,
		SortOrder:      filter.SortOrder,
		SkipUserFilter: isSuperAdmin,
	}

	if filter.CategoryID != "" {
		catID, err := uuid.Parse(filter.CategoryID)
		if err == nil {
			params.CategoryID = &catID
		}
	}

	if filter.UnitID != "" {
		unitID, err := uuid.Parse(filter.UnitID)
		if err == nil {
			params.UnitID = &unitID
		}
	}

	// For super admins, skip tenant scope to see all products
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

	result, err := h.productService.ListProducts(ctx, *userID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Products retrieved successfully", result)
}

// listWithCursor handles listing products with cursor-based pagination
func (h *ProductHandler) listWithCursor(c *gin.Context, userID uuid.UUID, isSuperAdmin bool) {
	var filter request.ProductFilterRequest
	if err := c.ShouldBindQuery(&filter); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	limit := 15
	if filter.Limit > 0 {
		limit = filter.Limit
	}
	cursor := c.Query("cursor")
	direction := c.DefaultQuery("direction", "next")

	params := &repository.ProductCursorFilterParams{
		Cursor: &pagination.CursorParams{
			Cursor:    cursor,
			Direction: pagination.CursorDirection(direction),
			Limit:     limit,
		},
		Search:         filter.Search,
		LowStock:       filter.LowStock,
		SkipUserFilter: isSuperAdmin,
	}

	if filter.CategoryID != "" {
		catID, err := uuid.Parse(filter.CategoryID)
		if err == nil {
			params.CategoryID = &catID
		}
	}

	if filter.UnitID != "" {
		unitID, err := uuid.Parse(filter.UnitID)
		if err == nil {
			params.UnitID = &unitID
		}
	}

	// For super admins, skip tenant scope to see all products
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

	result, err := h.productService.ListProductsWithCursor(ctx, userID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, 200, "Products retrieved successfully", result)
}

// Create handles creating a product
func (h *ProductHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req request.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	product, err := h.productService.CreateProduct(c.Request.Context(), &service.CreateProductInput{
		UserID:        *userID,
		CategoryID:    req.CategoryID,
		UnitID:        req.UnitID,
		Name:          req.Name,
		Code:          req.Code,
		Quantity:      req.Quantity,
		QuantityAlert: req.QuantityAlert,
		BuyingPrice:   req.BuyingPrice,
		SellingPrice:  req.SellingPrice,
		Tax:           req.Tax,
		TaxType:       req.TaxType,
		Notes:         req.Notes,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Product created successfully", product)
}

// Get handles getting a single product
func (h *ProductHandler) Get(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "Product slug is required")
		return
	}

	product, err := h.productService.GetProduct(c.Request.Context(), slug)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Product retrieved successfully", product)
}

// Update handles updating a product
func (h *ProductHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "Product slug is required")
		return
	}

	var req request.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body: "+err.Error())
		return
	}

	product, err := h.productService.UpdateProduct(c.Request.Context(), &service.UpdateProductInput{
		UserID:        *userID,
		ProductSlug:   slug,
		SkipUserCheck: isSuperAdmin,
		CategoryID:    req.CategoryID,
		UnitID:        req.UnitID,
		Name:          req.Name,
		Code:          req.Code,
		Quantity:      req.Quantity,
		QuantityAlert: req.QuantityAlert,
		BuyingPrice:   req.BuyingPrice,
		SellingPrice:  req.SellingPrice,
		Tax:           req.Tax,
		TaxType:       req.TaxType,
		Notes:         req.Notes,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Product updated successfully", product)
}

// Delete handles deleting a product by slug
func (h *ProductHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "Product slug is required")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	if err := h.productService.DeleteProduct(c.Request.Context(), *userID, slug, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// GetLowStock handles getting low stock products
func (h *ProductHandler) GetLowStock(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	products, err := h.productService.GetLowStockProducts(c.Request.Context(), *userID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Low stock products retrieved successfully", products)
}

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List handles listing categories
func (h *CategoryHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)
	page := 1
	perPage := 50
	search := c.Query("search")

	params := &pagination.PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	// For super admins, skip tenant scope to see all categories
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

	result, err := h.categoryService.ListCategories(ctx, *userID, params, search, isSuperAdmin)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Categories retrieved successfully", result)
}

// Create handles creating a category
func (h *CategoryHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	category, err := h.categoryService.CreateCategory(c.Request.Context(), &service.CreateCategoryInput{
		UserID: *userID,
		Name:   req.Name,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Category created successfully", category)
}

// Update handles updating a category
func (h *CategoryHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	category, err := h.categoryService.UpdateCategory(c.Request.Context(), &service.UpdateCategoryInput{
		UserID:       *userID,
		ID:           id,
		IsSuperAdmin: isSuperAdmin,
		Name:         req.Name,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Category updated successfully", category)
}

// Delete handles deleting a category
func (h *CategoryHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid category ID")
		return
	}

	if err := h.categoryService.DeleteCategory(c.Request.Context(), *userID, id, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// UnitHandler handles unit-related HTTP requests
type UnitHandler struct {
	unitService *service.UnitService
}

// NewUnitHandler creates a new unit handler
func NewUnitHandler(unitService *service.UnitService) *UnitHandler {
	return &UnitHandler{unitService: unitService}
}

// List handles listing units
func (h *UnitHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)
	params := &pagination.PaginationParams{
		Page:    1,
		PerPage: 50,
	}
	search := c.Query("search")

	// For super admins, skip tenant scope to see all units
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

	result, err := h.unitService.ListUnits(ctx, *userID, params, search, isSuperAdmin)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Units retrieved successfully", result)
}

// Create handles creating a unit
func (h *UnitHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		ShortCode string `json:"short_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	unit, err := h.unitService.CreateUnit(c.Request.Context(), &service.CreateUnitInput{
		UserID:    *userID,
		Name:      req.Name,
		ShortCode: req.ShortCode,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Unit created successfully", unit)
}

// Update handles updating a unit
func (h *UnitHandler) Update(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid unit ID")
		return
	}

	var req struct {
		Name      string `json:"name" binding:"required"`
		ShortCode string `json:"short_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	unit, err := h.unitService.UpdateUnit(c.Request.Context(), &service.UpdateUnitInput{
		UserID:       *userID,
		ID:           id,
		IsSuperAdmin: isSuperAdmin,
		Name:         req.Name,
		ShortCode:    req.ShortCode,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Unit updated successfully", unit)
}

// Delete handles deleting a unit
func (h *UnitHandler) Delete(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid unit ID")
		return
	}

	if err := h.unitService.DeleteUnit(c.Request.Context(), *userID, id, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// Suppress unused import warning
var _ = enum.OrderStatusPending
