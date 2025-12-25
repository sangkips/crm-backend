package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// CustomerRepository defines the interface for customer data operations
type CustomerRepository interface {
	Create(ctx context.Context, customer *entity.Customer) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error)
	GetByEmail(ctx context.Context, email string) (*entity.Customer, error)
	Update(ctx context.Context, customer *entity.Customer) error
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns customers with page-based pagination. If skipUserFilter is true, returns all customers.
	List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Customer, int64, error)
	// ListWithCursor returns customers using cursor-based pagination. If skipUserFilter is true, returns all customers.
	ListWithCursor(ctx context.Context, userID uuid.UUID, params *pagination.CursorParams, search string, skipUserFilter bool) ([]entity.Customer, error)
}

// SupplierRepository defines the interface for supplier data operations
type SupplierRepository interface {
	Create(ctx context.Context, supplier *entity.Supplier) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error)
	GetByEmail(ctx context.Context, email string) (*entity.Supplier, error)
	Update(ctx context.Context, supplier *entity.Supplier) error
	Delete(ctx context.Context, id uuid.UUID) error
	// List returns suppliers. If skipUserFilter is true, returns all suppliers.
	List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Supplier, int64, error)
}
