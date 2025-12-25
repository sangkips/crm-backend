package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
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
