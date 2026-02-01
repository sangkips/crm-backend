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

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) domainRepo.OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).
		Scopes(TenantScope(ctx)).
		Preload("Customer").
		First(&order, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &order, err
}

func (r *orderRepository) GetByInvoiceNo(ctx context.Context, invoiceNo string) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).Scopes(TenantScope(ctx)).First(&order, "invoice_no = ?", invoiceNo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &order, err
}

func (r *orderRepository) Update(ctx context.Context, order *entity.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Order{}, "id = ?", id).Error
}

func (r *orderRepository) List(ctx context.Context, userID uuid.UUID, params *domainRepo.OrderFilterParams) ([]entity.Order, int64, error) {
	var orders []entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{}).Scopes(TenantScope(ctx))
	if !params.SkipUserFilter && userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}

	if params.Search != "" {
		query = query.Where("invoice_no ILIKE ?", "%"+params.Search+"%")
	}

	if params.Status != nil {
		query = query.Where("order_status = ?", *params.Status)
	}

	if params.CustomerID != nil {
		query = query.Where("customer_id = ?", *params.CustomerID)
	}

	if params.StartDate != nil {
		query = query.Where("order_date >= ?", *params.StartDate)
	}

	if params.EndDate != nil {
		query = query.Where("order_date <= ?", *params.EndDate)
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
		Find(&orders).Error

	return orders, total, err
}

func (r *orderRepository) GetWithDetails(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	var order entity.Order
	err := r.db.WithContext(ctx).
		Scopes(TenantScope(ctx)).
		Preload("Customer").
		Preload("Details.Product").
		Preload("Details.Product.Category").
		First(&order, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &order, err
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status enum.OrderStatus) error {
	return r.db.WithContext(ctx).Model(&entity.Order{}).
		Where("id = ?", id).
		Update("order_status", status).Error
}

func (r *orderRepository) GetDueOrders(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) ([]entity.Order, int64, error) {
	var orders []entity.Order
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Order{}).Scopes(TenantScope(ctx)).
		Where("due > 0")
	if userID != uuid.Nil {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).
		Preload("Customer").
		Order("created_at DESC").
		Find(&orders).Error

	return orders, total, err
}

// ListWithCursor returns orders using cursor-based pagination
func (r *orderRepository) ListWithCursor(ctx context.Context, userID uuid.UUID, params *domainRepo.OrderCursorFilterParams) ([]entity.Order, error) {
	var orders []entity.Order

	params.Cursor.Validate()
	query := r.db.WithContext(ctx).Model(&entity.Order{})
	if !params.SkipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if params.Search != "" {
		query = query.Where("invoice_no ILIKE ?", "%"+params.Search+"%")
	}

	if params.Status != nil {
		query = query.Where("order_status = ?", *params.Status)
	}

	if params.CustomerID != nil {
		query = query.Where("customer_id = ?", *params.CustomerID)
	}

	if params.StartDate != nil {
		query = query.Where("order_date >= ?", *params.StartDate)
	}

	if params.EndDate != nil {
		query = query.Where("order_date <= ?", *params.EndDate)
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
		Preload("Customer").
		Order("created_at ASC, id ASC").
		Find(&orders).Error

	return orders, err
}

type orderDetailRepository struct {
	db *gorm.DB
}

// NewOrderDetailRepository creates a new order detail repository
func NewOrderDetailRepository(db *gorm.DB) domainRepo.OrderDetailRepository {
	return &orderDetailRepository{db: db}
}

func (r *orderDetailRepository) Create(ctx context.Context, detail *entity.OrderDetail) error {
	return r.db.WithContext(ctx).Create(detail).Error
}

func (r *orderDetailRepository) CreateBatch(ctx context.Context, details []entity.OrderDetail) error {
	return r.db.WithContext(ctx).Create(&details).Error
}

func (r *orderDetailRepository) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.OrderDetail, error) {
	var details []entity.OrderDetail
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("order_id = ?", orderID).
		Find(&details).Error
	return details, err
}

func (r *orderDetailRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.OrderDetail{}, "id = ?", id).Error
}

func (r *orderDetailRepository) DeleteByOrderID(ctx context.Context, orderID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.OrderDetail{}, "order_id = ?", orderID).Error
}
