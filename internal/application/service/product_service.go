package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/pagination"
	"github.com/sangkips/investify-api/pkg/utils"
)

// ProductService handles product-related operations
type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
	unitRepo     repository.UnitRepository
}

// NewProductService creates a new product service
func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	unitRepo repository.UnitRepository,
) *ProductService {
	return &ProductService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
		unitRepo:     unitRepo,
	}
}

// CreateProductInput represents the create product input
type CreateProductInput struct {
	UserID        uuid.UUID
	CategoryID    *uuid.UUID
	UnitID        *uuid.UUID
	Name          string
	Code          string
	Quantity      int
	QuantityAlert int
	BuyingPrice   float64
	SellingPrice  float64
	Tax           int
	TaxType       int
	Notes         *string
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(ctx context.Context, input *CreateProductInput) (*entity.Product, error) {
	// Extract tenant ID from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	// Auto-generate code if not provided
	code := input.Code
	if code == "" {
		code = utils.GenerateProductCode()
	}

	// Check if code already exists
	existingProduct, err := s.productRepo.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	if existingProduct != nil {
		return nil, apperror.NewConflictError("Product code already exists")
	}

	// Generate slug
	slug := utils.Slugify(input.Name)

	product := &entity.Product{
		TenantID:      tenantID,
		UserID:        input.UserID,
		CategoryID:    input.CategoryID,
		UnitID:        input.UnitID,
		Name:          input.Name,
		Slug:          slug,
		Code:          code,
		Quantity:      input.Quantity,
		QuantityAlert: input.QuantityAlert,
		Tax:           input.Tax,
		Notes:         input.Notes,
	}
	product.SetBuyingPriceFromDecimal(input.BuyingPrice)
	product.SetSellingPriceFromDecimal(input.SellingPrice)

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return s.productRepo.GetByID(ctx, product.ID)
}

// GetProduct retrieves a product by slug
func (s *ProductService) GetProduct(ctx context.Context, slug string) (*entity.Product, error) {
	product, err := s.productRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, apperror.NewNotFoundError("Product")
	}
	return product, nil
}

// GetProductByID retrieves a product by ID
func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	product, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, apperror.NewNotFoundError("Product")
	}
	return product, nil
}

// ListProducts lists products with filtering
func (s *ProductService) ListProducts(ctx context.Context, userID uuid.UUID, params *repository.ProductFilterParams) (*pagination.PaginatedResult[entity.Product], error) {
	products, total, err := s.productRepo.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Pagination.Page, params.Pagination.PerPage, total)
	return pagination.NewPaginatedResult(products, pag), nil
}

// ListProductsWithCursor lists products with cursor-based pagination
func (s *ProductService) ListProductsWithCursor(ctx context.Context, userID uuid.UUID, params *repository.ProductCursorFilterParams) (*pagination.CursorPaginatedResult[entity.Product], error) {
	products, err := s.productRepo.ListWithCursor(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	hasPrev := params.Cursor.Cursor != ""

	cursorPag, items := pagination.NewCursorPagination(products, params.Cursor.Limit,
		func(p entity.Product) string { return p.ID.String() },
		func(p entity.Product) time.Time { return p.CreatedAt },
	)
	cursorPag.HasPrev = hasPrev

	return pagination.NewCursorPaginatedResult(items, cursorPag), nil
}

// UpdateProductInput represents the update product input
type UpdateProductInput struct {
	UserID        uuid.UUID
	ProductSlug   string
	SkipUserCheck bool // If true (super-admin), skip ownership check
	CategoryID    *uuid.UUID
	UnitID        *uuid.UUID
	Name          *string
	Code          *string
	Quantity      *int
	QuantityAlert *int
	BuyingPrice   *float64
	SellingPrice  *float64
	Tax           *int
	TaxType       *int
	Notes         *string
}

// UpdateProduct updates a product
func (s *ProductService) UpdateProduct(ctx context.Context, input *UpdateProductInput) (*entity.Product, error) {
	product, err := s.productRepo.GetBySlug(ctx, input.ProductSlug)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, apperror.NewNotFoundError("Product")
	}

	// Ensure user owns the product (unless super-admin)
	if !input.SkipUserCheck && product.UserID != input.UserID {
		return nil, apperror.ErrForbidden
	}

	// Check if new code is unique
	if input.Code != nil && *input.Code != product.Code {
		existingProduct, err := s.productRepo.GetByCode(ctx, *input.Code)
		if err != nil {
			return nil, err
		}
		if existingProduct != nil && existingProduct.ID != product.ID {
			return nil, apperror.NewConflictError("Product code already exists")
		}
		product.Code = *input.Code
	}

	if input.CategoryID != nil {
		product.CategoryID = input.CategoryID
	}
	if input.UnitID != nil {
		product.UnitID = input.UnitID
	}
	if input.Name != nil {
		product.Name = *input.Name
		product.Slug = utils.Slugify(*input.Name)
	}
	if input.Quantity != nil {
		product.Quantity = *input.Quantity
	}
	if input.QuantityAlert != nil {
		product.QuantityAlert = *input.QuantityAlert
	}
	if input.BuyingPrice != nil {
		product.SetBuyingPriceFromDecimal(*input.BuyingPrice)
	}
	if input.SellingPrice != nil {
		product.SetSellingPriceFromDecimal(*input.SellingPrice)
	}
	if input.Tax != nil {
		product.Tax = *input.Tax
	}
	if input.TaxType != nil {
		product.TaxType = enum.TaxType(*input.TaxType)
	}
	if input.Notes != nil {
		product.Notes = input.Notes
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}

	return s.productRepo.GetByID(ctx, product.ID)
}

// DeleteProduct deletes a product
// If skipOwnerCheck is true (e.g., for super-admins), ownership check is bypassed
func (s *ProductService) DeleteProduct(ctx context.Context, userID uuid.UUID, slug string, skipOwnerCheck bool) error {
	product, err := s.productRepo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}
	if product == nil {
		return apperror.NewNotFoundError("Product")
	}

	// Only check ownership if not a super-admin
	if !skipOwnerCheck && product.UserID != userID {
		return apperror.ErrForbidden
	}

	return s.productRepo.Delete(ctx, product.ID)
}

// GetLowStockProducts returns products with low stock
func (s *ProductService) GetLowStockProducts(ctx context.Context, userID uuid.UUID) ([]entity.Product, error) {
	return s.productRepo.GetLowStock(ctx, userID)
}

// ImportProductRow represents a single row from the import file
type ImportProductRow struct {
	Name          string
	Code          string
	Quantity      int
	QuantityAlert int
	BuyingPrice   float64
	SellingPrice  float64
	Tax           int
	TaxType       int
	Notes         string
	CategoryName  string
	UnitName      string
}

// ImportResult contains the result of a product import operation
type ImportResult struct {
	TotalRows  int              `json:"total_rows"`
	Successful int              `json:"successful"`
	Failed     int              `json:"failed"`
	Errors     []ImportRowError `json:"errors,omitempty"`
}

// ImportRowError describes an error for a specific row during import
type ImportRowError struct {
	Row     int    `json:"row"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ImportProducts validates and bulk-creates products from parsed import rows
func (s *ProductService) ImportProducts(ctx context.Context, userID uuid.UUID, rows []ImportProductRow) (*ImportResult, error) {
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	result := &ImportResult{TotalRows: len(rows)}
	var rowErrors []ImportRowError

	// Load categories and units for the tenant for name-based matching
	categoryMap := make(map[string]*uuid.UUID)
	unitMap := make(map[string]*uuid.UUID)

	categories, _, _ := s.categoryRepo.List(ctx, uuid.Nil, &pagination.PaginationParams{Page: 1, PerPage: 1000}, "", true)
	for i := range categories {
		categoryMap[strings.ToLower(categories[i].Name)] = &categories[i].ID
	}

	units, _, _ := s.unitRepo.List(ctx, uuid.Nil, &pagination.PaginationParams{Page: 1, PerPage: 1000}, "", true)
	for i := range units {
		unitMap[strings.ToLower(units[i].Name)] = &units[i].ID
	}

	// Track codes seen in this import batch to detect duplicates within the file
	seenCodes := make(map[string]int) // code -> row number (1-indexed)

	var validProducts []entity.Product

	for i, row := range rows {
		rowNum := i + 2 // +2 because row 1 is the header, data starts at row 2

		// Validate required fields
		if strings.TrimSpace(row.Name) == "" {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Field: "name", Message: "Name is required"})
			continue
		}

		// Auto-generate code if empty
		code := strings.TrimSpace(row.Code)
		if code == "" {
			code = utils.GenerateProductCode()
		}

		// Check for duplicate code within the file
		if prevRow, exists := seenCodes[code]; exists {
			rowErrors = append(rowErrors, ImportRowError{
				Row:     rowNum,
				Field:   "code",
				Message: fmt.Sprintf("Duplicate code '%s' (same as row %d)", code, prevRow),
			})
			continue
		}

		// Check if code already exists in DB
		existingProduct, err := s.productRepo.GetByCode(ctx, code)
		if err != nil {
			rowErrors = append(rowErrors, ImportRowError{Row: rowNum, Field: "code", Message: "Error checking code: " + err.Error()})
			continue
		}
		if existingProduct != nil {
			rowErrors = append(rowErrors, ImportRowError{
				Row:     rowNum,
				Field:   "code",
				Message: fmt.Sprintf("Product code '%s' already exists", code),
			})
			continue
		}

		seenCodes[code] = rowNum

		// Generate slug with uniqueness suffix
		slug := utils.Slugify(row.Name) + "-" + strings.ToLower(uuid.New().String()[:8])

		// Match category by name
		var categoryID *uuid.UUID
		if row.CategoryName != "" {
			if id, ok := categoryMap[strings.ToLower(strings.TrimSpace(row.CategoryName))]; ok {
				categoryID = id
			}
		}

		// Match unit by name
		var unitID *uuid.UUID
		if row.UnitName != "" {
			if id, ok := unitMap[strings.ToLower(strings.TrimSpace(row.UnitName))]; ok {
				unitID = id
			}
		}

		product := entity.Product{
			TenantID:      tenantID,
			UserID:        userID,
			CategoryID:    categoryID,
			UnitID:        unitID,
			Name:          strings.TrimSpace(row.Name),
			Slug:          slug,
			Code:          code,
			Quantity:      row.Quantity,
			QuantityAlert: row.QuantityAlert,
			Tax:           row.Tax,
			TaxType:       enum.TaxType(row.TaxType),
		}
		product.SetBuyingPriceFromDecimal(row.BuyingPrice)
		product.SetSellingPriceFromDecimal(row.SellingPrice)

		if row.Notes != "" {
			notes := row.Notes
			product.Notes = &notes
		}

		validProducts = append(validProducts, product)
	}

	// Batch create valid products
	if len(validProducts) > 0 {
		if err := s.productRepo.CreateBatch(ctx, validProducts); err != nil {
			return nil, apperror.NewAppError(500, "Failed to import products: "+err.Error())
		}
	}

	result.Successful = len(validProducts)
	result.Failed = len(rowErrors)
	result.Errors = rowErrors

	return result, nil
}
