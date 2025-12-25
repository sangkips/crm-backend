package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	domainRepo "github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/pagination"
	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *gorm.DB) domainRepo.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *entity.Customer) error {
	return r.db.WithContext(ctx).Create(customer).Error
}

func (r *customerRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).First(&customer, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &customer, err
}

func (r *customerRepository) GetByEmail(ctx context.Context, email string) (*entity.Customer, error) {
	var customer entity.Customer
	err := r.db.WithContext(ctx).First(&customer, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &customer, err
}

func (r *customerRepository) Update(ctx context.Context, customer *entity.Customer) error {
	return r.db.WithContext(ctx).Save(customer).Error
}

func (r *customerRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Customer{}, "id = ?", id).Error
}

func (r *customerRepository) List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Customer, int64, error) {
	var customers []entity.Customer
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Customer{})
	if !skipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).
		Order("name ASC").
		Find(&customers).Error

	return customers, total, err
}

// ListWithCursor returns customers using cursor-based pagination
// Fetches limit+1 items to detect if there are more results
func (r *customerRepository) ListWithCursor(ctx context.Context, userID uuid.UUID, params *pagination.CursorParams, search string, skipUserFilter bool) ([]entity.Customer, error) {
	var customers []entity.Customer

	params.Validate()
	query := r.db.WithContext(ctx).Model(&entity.Customer{})
	if !skipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Decode cursor if provided
	cursor, err := params.DecodeCursor()
	if err != nil {
		return nil, err
	}

	// Apply cursor-based filtering using created_at and id
	if cursor != nil {
		if params.Direction == pagination.CursorDirectionNext {
			// For next page: get items created after the cursor, ordered by created_at, id
			query = query.Where("(created_at, id) > (?, ?)", cursor.CreatedAt, cursor.ID)
		} else {
			// For prev page: get items created before the cursor
			query = query.Where("(created_at, id) < (?, ?)", cursor.CreatedAt, cursor.ID)
		}
	}

	// Fetch limit+1 to detect hasMore
	err = query.Limit(params.Limit + 1).
		Order("created_at ASC, id ASC").
		Find(&customers).Error

	return customers, err
}

type supplierRepository struct {
	db *gorm.DB
}

// NewSupplierRepository creates a new supplier repository
func NewSupplierRepository(db *gorm.DB) domainRepo.SupplierRepository {
	return &supplierRepository{db: db}
}

func (r *supplierRepository) Create(ctx context.Context, supplier *entity.Supplier) error {
	return r.db.WithContext(ctx).Create(supplier).Error
}

func (r *supplierRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Supplier, error) {
	var supplier entity.Supplier
	err := r.db.WithContext(ctx).First(&supplier, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &supplier, err
}

func (r *supplierRepository) GetByEmail(ctx context.Context, email string) (*entity.Supplier, error) {
	var supplier entity.Supplier
	err := r.db.WithContext(ctx).First(&supplier, "email = ?", email).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &supplier, err
}

func (r *supplierRepository) Update(ctx context.Context, supplier *entity.Supplier) error {
	return r.db.WithContext(ctx).Save(supplier).Error
}

func (r *supplierRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Supplier{}, "id = ?", id).Error
}

func (r *supplierRepository) List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Supplier, int64, error) {
	var suppliers []entity.Supplier
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Supplier{})
	if !skipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if search != "" {
		query = query.Where("name ILIKE ? OR email ILIKE ? OR phone ILIKE ? OR shopname ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).
		Order("name ASC").
		Find(&suppliers).Error

	return suppliers, total, err
}
