package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
	"gorm.io/gorm"
)

type quotationRepository struct {
	db *gorm.DB
}

// NewQuotationRepository creates a new quotation repository
func NewQuotationRepository(db *gorm.DB) domainRepo.QuotationRepository {
	return &quotationRepository{db: db}
}

func (r *quotationRepository) Create(ctx context.Context, quotation *entity.Quotation) error {
	return r.db.WithContext(ctx).Create(quotation).Error
}

func (r *quotationRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Quotation, error) {
	var quotation entity.Quotation
	err := r.db.WithContext(ctx).
		Preload("Customer").
		First(&quotation, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &quotation, err
}

func (r *quotationRepository) GetByReference(ctx context.Context, reference string) (*entity.Quotation, error) {
	var quotation entity.Quotation
	err := r.db.WithContext(ctx).First(&quotation, "reference = ?", reference).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &quotation, err
}

func (r *quotationRepository) Update(ctx context.Context, quotation *entity.Quotation) error {
	return r.db.WithContext(ctx).Save(quotation).Error
}

func (r *quotationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Quotation{}, "id = ?", id).Error
}

func (r *quotationRepository) List(ctx context.Context, userID uuid.UUID, params *domainRepo.QuotationFilterParams) ([]entity.Quotation, int64, error) {
	var quotations []entity.Quotation
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Quotation{})

	// Only filter by user_id if a non-zero userID is provided (super-admin can see all)
	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}

	if params.Search != "" {
		query = query.Where("reference ILIKE ? OR customer_name ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if params.CustomerID != nil {
		query = query.Where("customer_id = ?", *params.CustomerID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Sorting
	sortBy := "created_at"
	sortOrder := "DESC"
	if params.SortBy != "" {
		sortBy = params.SortBy
	}
	if params.SortOrder != "" && (params.SortOrder == "ASC" || params.SortOrder == "asc") {
		sortOrder = "ASC"
	}

	params.Pagination.Validate()
	err := query.Offset(params.Pagination.Offset()).Limit(params.Pagination.PerPage).
		Preload("Customer").
		Order(sortBy + " " + sortOrder).
		Find(&quotations).Error

	return quotations, total, err
}

func (r *quotationRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*entity.Quotation, error) {
	var quotation entity.Quotation
	err := r.db.WithContext(ctx).
		Preload("Customer").
		Preload("Details.Product").
		First(&quotation, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &quotation, err
}

func (r *quotationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status enum.QuotationStatus) error {
	return r.db.WithContext(ctx).Model(&entity.Quotation{}).
		Where("id = ?", id).
		Update("status", status).Error
}

func (r *quotationRepository) GetNextReferenceNumber(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Quotation{}).Count(&count).Error
	return int(count) + 1, err
}

type quotationDetailRepository struct {
	db *gorm.DB
}

// NewQuotationDetailRepository creates a new quotation detail repository
func NewQuotationDetailRepository(db *gorm.DB) domainRepo.QuotationDetailRepository {
	return &quotationDetailRepository{db: db}
}

func (r *quotationDetailRepository) Create(ctx context.Context, detail *entity.QuotationDetail) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

func (r *quotationDetailRepository) CreateBatch(ctx context.Context, details []entity.QuotationDetail) error {
	return r.db.WithContext(ctx).Create(&details).Error
}

func (r *quotationDetailRepository) GetByQuotationID(ctx context.Context, quotationID uuid.UUID) ([]entity.QuotationDetail, error) {
	var details []entity.QuotationDetail
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("quotation_id = ?", quotationID).
		Find(&details).Error
	return details, err
}

func (r *quotationDetailRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.QuotationDetail{}, "id = ?", id).Error
}

func (r *quotationDetailRepository) DeleteByQuotationID(ctx context.Context, quotationID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.QuotationDetail{}, "quotation_id = ?", quotationID).Error
}
