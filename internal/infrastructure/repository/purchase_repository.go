package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/pagination"
	"gorm.io/gorm"
)

type purchaseRepository struct {
	db *gorm.DB
}

// NewPurchaseRepository creates a new purchase repository
func NewPurchaseRepository(db *gorm.DB) domainRepo.PurchaseRepository {
	return &purchaseRepository{db: db}
}

func (r *purchaseRepository) Create(ctx context.Context, purchase *entity.Purchase) error {
	return r.db.WithContext(ctx).Create(purchase).Error
}

func (r *purchaseRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Purchase, error) {
	var purchase entity.Purchase
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("CreatedBy").
		First(&purchase, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &purchase, err
}

func (r *purchaseRepository) GetByPurchaseNo(ctx context.Context, purchaseNo string) (*entity.Purchase, error) {
	var purchase entity.Purchase
	err := r.db.WithContext(ctx).First(&purchase, "purchase_no = ?", purchaseNo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &purchase, err
}

func (r *purchaseRepository) Update(ctx context.Context, purchase *entity.Purchase) error {
	return r.db.WithContext(ctx).Save(purchase).Error
}

func (r *purchaseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Purchase{}, "id = ?", id).Error
}

func (r *purchaseRepository) List(ctx context.Context, userID uuid.UUID, params *domainRepo.PurchaseFilterParams) ([]entity.Purchase, int64, error) {
	var purchases []entity.Purchase
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Purchase{})
	if !params.SkipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if params.Search != "" {
		query = query.Where("purchase_no ILIKE ?", "%"+params.Search+"%")
	}

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if params.SupplierID != nil {
		query = query.Where("supplier_id = ?", *params.SupplierID)
	}

	if params.StartDate != nil {
		query = query.Where("date >= ?", *params.StartDate)
	}

	if params.EndDate != nil {
		query = query.Where("date <= ?", *params.EndDate)
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
		Preload("Supplier").
		Order(sortBy + " " + sortOrder).
		Find(&purchases).Error

	return purchases, total, err
}

func (r *purchaseRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*entity.Purchase, error) {
	var purchase entity.Purchase
	err := r.db.WithContext(ctx).
		Preload("Supplier").
		Preload("CreatedBy").
		Preload("UpdatedBy").
		Preload("Details.Product").
		First(&purchase, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &purchase, err
}

func (r *purchaseRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status enum.PurchaseStatus, updatedBy uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entity.Purchase{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_by": updatedBy,
		}).Error
}

func (r *purchaseRepository) GetPendingPurchases(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) ([]entity.Purchase, int64, error) {
	var purchases []entity.Purchase
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Purchase{}).
		Where("user_id = ? AND status = ?", userID, enum.PurchaseStatusPending)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).
		Preload("Supplier").
		Order("created_at DESC").
		Find(&purchases).Error

	return purchases, total, err
}

// ListWithCursor returns purchases using cursor-based pagination
func (r *purchaseRepository) ListWithCursor(ctx context.Context, userID uuid.UUID, params *domainRepo.PurchaseCursorFilterParams) ([]entity.Purchase, error) {
	var purchases []entity.Purchase

	params.Cursor.Validate()
	query := r.db.WithContext(ctx).Model(&entity.Purchase{})
	if !params.SkipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if params.Search != "" {
		query = query.Where("purchase_no ILIKE ?", "%"+params.Search+"%")
	}

	if params.Status != nil {
		query = query.Where("status = ?", *params.Status)
	}

	if params.SupplierID != nil {
		query = query.Where("supplier_id = ?", *params.SupplierID)
	}

	if params.StartDate != nil {
		query = query.Where("date >= ?", *params.StartDate)
	}

	if params.EndDate != nil {
		query = query.Where("date <= ?", *params.EndDate)
	}

	cursor, err := params.Cursor.DecodeCursor()
	if err != nil {
		return nil, err
	}

	if cursor != nil {
		if params.Cursor.Direction == pagination.CursorDirectionNext {
			query = query.Where("(created_at, id) > (?, ?)", cursor.CreatedAt, cursor.ID)
		} else {
			query = query.Where("(created_at, id) < (?, ?)", cursor.CreatedAt, cursor.ID)
		}
	}

	err = query.Limit(params.Cursor.Limit + 1).
		Preload("Supplier").
		Order("created_at ASC, id ASC").
		Find(&purchases).Error

	return purchases, err
}

type purchaseDetailRepository struct {
	db *gorm.DB
}

// NewPurchaseDetailRepository creates a new purchase detail repository
func NewPurchaseDetailRepository(db *gorm.DB) domainRepo.PurchaseDetailRepository {
	return &purchaseDetailRepository{db: db}
}

func (r *purchaseDetailRepository) Create(ctx context.Context, detail *entity.PurchaseDetail) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

func (r *purchaseDetailRepository) CreateBatch(ctx context.Context, details []entity.PurchaseDetail) error {
	return r.db.WithContext(ctx).Create(&details).Error
}

func (r *purchaseDetailRepository) GetByPurchaseID(ctx context.Context, purchaseID uuid.UUID) ([]entity.PurchaseDetail, error) {
	var details []entity.PurchaseDetail
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("purchase_id = ?", purchaseID).
		Find(&details).Error
	return details, err
}

func (r *purchaseDetailRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.PurchaseDetail{}, "id = ?", id).Error
}

func (r *purchaseDetailRepository) DeleteByPurchaseID(ctx context.Context, purchaseID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.PurchaseDetail{}, "purchase_id = ?", purchaseID).Error
}
