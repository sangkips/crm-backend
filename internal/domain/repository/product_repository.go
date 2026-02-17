package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// ProductRepository defines the interface for product data operations
type ProductRepository interface {
	Create(ctx context.Context, product *entity.Product) error
	CreateBatch(ctx context.Context, products []entity.Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	// GetByIDs retrieves multiple products by their IDs in a single query (prevents N+1)
	GetByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Product, error)
	GetByCode(ctx context.Context, code string) (*entity.Product, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, params *ProductFilterParams) ([]entity.Product, int64, error)
	ListWithCursor(ctx context.Context, userID uuid.UUID, params *ProductCursorFilterParams) ([]entity.Product, error)
	GetLowStock(ctx context.Context, userID uuid.UUID) ([]entity.Product, error)
	UpdateQuantity(ctx context.Context, id uuid.UUID, quantity int) error
	// UpdateQuantityBatch updates quantities for multiple products in a batch
	UpdateQuantityBatch(ctx context.Context, updates map[uuid.UUID]int) error
	// AtomicDecrementQuantity atomically decrements stock only if sufficient.
	// Returns (true, nil) if successful, (false, nil) if insufficient stock, (false, err) on error.
	AtomicDecrementQuantity(ctx context.Context, id uuid.UUID, amount int) (bool, error)
	// AtomicDecrementBatch atomically decrements stock for multiple products.
	// Returns map of product IDs that failed (insufficient stock) and any error.
	// If any product fails, the entire transaction is rolled back.
	AtomicDecrementBatch(ctx context.Context, decrements map[uuid.UUID]int) (failedIDs []uuid.UUID, err error)
	// AtomicIncrementBatch atomically increments stock for multiple products (for cancellations/returns).
	AtomicIncrementBatch(ctx context.Context, increments map[uuid.UUID]int) error
}

// ProductFilterParams contains filtering parameters for product queries
type ProductFilterParams struct {
	Pagination     *pagination.PaginationParams
	Search         string
	CategoryID     *uuid.UUID
	UnitID         *uuid.UUID
	LowStock       bool
	SortBy         string
	SortOrder      string
	SkipUserFilter bool // If true, returns all products (for super-admin)
}

// ProductCursorFilterParams contains cursor-based filtering parameters for product queries
type ProductCursorFilterParams struct {
	Cursor         *pagination.CursorParams
	Search         string
	CategoryID     *uuid.UUID
	UnitID         *uuid.UUID
	LowStock       bool
	SkipUserFilter bool // If true, returns all products (for super-admin)
}

// CategoryRepository defines the interface for category data operations
type CategoryRepository interface {
	Create(ctx context.Context, category *entity.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Category, error)
	Update(ctx context.Context, category *entity.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Category, int64, error)
}

// UnitRepository defines the interface for unit data operations
type UnitRepository interface {
	Create(ctx context.Context, unit *entity.Unit) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Unit, error)
	GetBySlug(ctx context.Context, slug string) (*entity.Unit, error)
	Update(ctx context.Context, unit *entity.Unit) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Unit, int64, error)
}
