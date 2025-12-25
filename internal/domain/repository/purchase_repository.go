package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// PurchaseRepository defines the interface for purchase data operations
type PurchaseRepository interface {
	Create(ctx context.Context, purchase *entity.Purchase) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Purchase, error)
	GetByPurchaseNo(ctx context.Context, purchaseNo string) (*entity.Purchase, error)
	Update(ctx context.Context, purchase *entity.Purchase) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, params *PurchaseFilterParams) ([]entity.Purchase, int64, error)
	ListWithCursor(ctx context.Context, userID uuid.UUID, params *PurchaseCursorFilterParams) ([]entity.Purchase, error)
	GetWithDetails(ctx context.Context, id uuid.UUID) (*entity.Purchase, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status enum.PurchaseStatus, updatedBy uuid.UUID) error
	GetPendingPurchases(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) ([]entity.Purchase, int64, error)
}

// PurchaseFilterParams contains filtering parameters for purchase queries
type PurchaseFilterParams struct {
	Pagination     *pagination.PaginationParams
	Search         string
	Status         *enum.PurchaseStatus
	SupplierID     *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	SortBy         string
	SortOrder      string
	SkipUserFilter bool // If true, returns all purchases (for super-admin)
}

// PurchaseCursorFilterParams contains cursor-based filtering for purchase queries
type PurchaseCursorFilterParams struct {
	Cursor         *pagination.CursorParams
	Search         string
	Status         *enum.PurchaseStatus
	SupplierID     *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	SkipUserFilter bool // If true, returns all purchases (for super-admin)
}

// PurchaseDetailRepository defines the interface for purchase detail data operations
type PurchaseDetailRepository interface {
	Create(ctx context.Context, detail *entity.PurchaseDetail) error
	CreateBatch(ctx context.Context, details []entity.PurchaseDetail) error
	GetByPurchaseID(ctx context.Context, purchaseID uuid.UUID) ([]entity.PurchaseDetail, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByPurchaseID(ctx context.Context, purchaseID uuid.UUID) error
}
