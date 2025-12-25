package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// OrderRepository defines the interface for order data operations
type OrderRepository interface {
	Create(ctx context.Context, order *entity.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	GetByInvoiceNo(ctx context.Context, invoiceNo string) (*entity.Order, error)
	Update(ctx context.Context, order *entity.Order) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, params *OrderFilterParams) ([]entity.Order, int64, error)
	ListWithCursor(ctx context.Context, userID uuid.UUID, params *OrderCursorFilterParams) ([]entity.Order, error)
	GetWithDetails(ctx context.Context, id uuid.UUID) (*entity.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status enum.OrderStatus) error
	GetDueOrders(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) ([]entity.Order, int64, error)
}

// OrderFilterParams contains filtering parameters for order queries
type OrderFilterParams struct {
	Pagination     *pagination.PaginationParams
	Search         string
	Status         *enum.OrderStatus
	CustomerID     *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	SortBy         string
	SortOrder      string
	SkipUserFilter bool // If true, returns all orders (for super-admin)
}

// OrderCursorFilterParams contains cursor-based filtering for order queries
type OrderCursorFilterParams struct {
	Cursor         *pagination.CursorParams
	Search         string
	Status         *enum.OrderStatus
	CustomerID     *uuid.UUID
	StartDate      *time.Time
	EndDate        *time.Time
	SkipUserFilter bool // If true, returns all orders (for super-admin)
}

// OrderDetailRepository defines the interface for order detail data operations
type OrderDetailRepository interface {
	Create(ctx context.Context, detail *entity.OrderDetail) error
	CreateBatch(ctx context.Context, details []entity.OrderDetail) error
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.OrderDetail, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error
}
