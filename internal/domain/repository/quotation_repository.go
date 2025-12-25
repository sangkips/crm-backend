package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// QuotationRepository defines the interface for quotation data operations
type QuotationRepository interface {
	Create(ctx context.Context, quotation *entity.Quotation) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Quotation, error)
	GetByReference(ctx context.Context, reference string) (*entity.Quotation, error)
	Update(ctx context.Context, quotation *entity.Quotation) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, userID uuid.UUID, params *QuotationFilterParams) ([]entity.Quotation, int64, error)
	GetWithDetails(ctx context.Context, id uuid.UUID) (*entity.Quotation, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status enum.QuotationStatus) error
	GetNextReferenceNumber(ctx context.Context) (int, error)
}

// QuotationFilterParams contains filtering parameters for quotation queries
type QuotationFilterParams struct {
	Pagination *pagination.PaginationParams
	Search     string
	Status     *enum.QuotationStatus
	CustomerID *uuid.UUID
	SortBy     string
	SortOrder  string
}

// QuotationDetailRepository defines the interface for quotation detail data operations
type QuotationDetailRepository interface {
	Create(ctx context.Context, detail *entity.QuotationDetail) error
	CreateBatch(ctx context.Context, details []entity.QuotationDetail) error
	GetByQuotationID(ctx context.Context, quotationID uuid.UUID) ([]entity.QuotationDetail, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByQuotationID(ctx context.Context, quotationID uuid.UUID) error
}
