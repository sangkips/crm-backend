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

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository creates a new product repository
func NewProductRepository(db *gorm.DB) domainRepo.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).
		Preload("Category").Preload("Unit").
		First(&product, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

func (r *productRepository) GetBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).
		Preload("Category").Preload("Unit").
		First(&product, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

// GetByIDs retrieves multiple products by their IDs in a single query
func (r *productRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]entity.Product, error) {
	if len(ids) == 0 {
		return []entity.Product{}, nil
	}
	var products []entity.Product
	err := r.db.WithContext(ctx).
		Preload("Category").Preload("Unit").
		Where("id IN ?", ids).
		Find(&products).Error
	return products, err
}

func (r *productRepository) GetByCode(ctx context.Context, code string) (*entity.Product, error) {
	var product entity.Product
	err := r.db.WithContext(ctx).First(&product, "code = ?", code).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &product, err
}

func (r *productRepository) Update(ctx context.Context, product *entity.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Product{}, "id = ?", id).Error
}

func (r *productRepository) List(ctx context.Context, userID uuid.UUID, params *domainRepo.ProductFilterParams) ([]entity.Product, int64, error) {
	var products []entity.Product
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Product{})
	if !params.SkipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if params.Search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	if params.CategoryID != nil {
		query = query.Where("category_id = ?", *params.CategoryID)
	}

	if params.UnitID != nil {
		query = query.Where("unit_id = ?", *params.UnitID)
	}

	if params.LowStock {
		query = query.Where("quantity <= quantity_alert")
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
		Preload("Category").Preload("Unit").
		Order(sortBy + " " + sortOrder).
		Find(&products).Error

	return products, total, err
}

func (r *productRepository) GetLowStock(ctx context.Context, userID uuid.UUID) ([]entity.Product, error) {
	var products []entity.Product
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND quantity <= quantity_alert", userID).
		Preload("Category").Preload("Unit").
		Find(&products).Error
	return products, err
}

func (r *productRepository) UpdateQuantity(ctx context.Context, id uuid.UUID, quantity int) error {
	return r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ?", id).
		Update("quantity", quantity).Error
}

// UpdateQuantityBatch updates quantities for multiple products in a single transaction
func (r *productRepository) UpdateQuantityBatch(ctx context.Context, updates map[uuid.UUID]int) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for id, quantity := range updates {
			if err := tx.Model(&entity.Product{}).
				Where("id = ?", id).
				Update("quantity", quantity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// AtomicDecrementQuantity atomically decrements stock only if sufficient quantity exists.
// Uses: UPDATE products SET quantity = quantity - amount WHERE id = ? AND quantity >= amount
func (r *productRepository) AtomicDecrementQuantity(ctx context.Context, id uuid.UUID, amount int) (bool, error) {
	result := r.db.WithContext(ctx).Model(&entity.Product{}).
		Where("id = ? AND quantity >= ?", id, amount).
		Update("quantity", gorm.Expr("quantity - ?", amount))

	if result.Error != nil {
		return false, result.Error
	}

	// If no rows were affected, insufficient stock
	return result.RowsAffected > 0, nil
}

// AtomicDecrementBatch atomically decrements stock for multiple products in a single transaction.
// If any product has insufficient stock, the entire transaction is rolled back.
func (r *productRepository) AtomicDecrementBatch(ctx context.Context, decrements map[uuid.UUID]int) ([]uuid.UUID, error) {
	if len(decrements) == 0 {
		return nil, nil
	}

	var failedIDs []uuid.UUID

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for id, amount := range decrements {
			result := tx.Model(&entity.Product{}).
				Where("id = ? AND quantity >= ?", id, amount).
				Update("quantity", gorm.Expr("quantity - ?", amount))

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				failedIDs = append(failedIDs, id)
			}
		}

		// If any products failed, rollback entire transaction
		if len(failedIDs) > 0 {
			return gorm.ErrInvalidTransaction
		}

		return nil
	})

	// If we rolled back due to insufficient stock, return the failed IDs without the transaction error
	if err == gorm.ErrInvalidTransaction && len(failedIDs) > 0 {
		return failedIDs, nil
	}

	return failedIDs, err
}

// AtomicIncrementBatch atomically increments stock for multiple products (for cancellations/returns).
func (r *productRepository) AtomicIncrementBatch(ctx context.Context, increments map[uuid.UUID]int) error {
	if len(increments) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for id, amount := range increments {
			if err := tx.Model(&entity.Product{}).
				Where("id = ?", id).
				Update("quantity", gorm.Expr("quantity + ?", amount)).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// ListWithCursor returns products using cursor-based pagination
func (r *productRepository) ListWithCursor(ctx context.Context, userID uuid.UUID, params *domainRepo.ProductCursorFilterParams) ([]entity.Product, error) {
	var products []entity.Product

	params.Cursor.Validate()
	query := r.db.WithContext(ctx).Model(&entity.Product{})
	if !params.SkipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if params.Search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?",
			"%"+params.Search+"%", "%"+params.Search+"%")
	}

	if params.CategoryID != nil {
		query = query.Where("category_id = ?", *params.CategoryID)
	}

	if params.UnitID != nil {
		query = query.Where("unit_id = ?", *params.UnitID)
	}

	if params.LowStock {
		query = query.Where("quantity <= quantity_alert")
	}

	// Decode cursor if provided
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

	// Fetch limit+1 to detect hasMore
	err = query.Limit(params.Cursor.Limit + 1).
		Preload("Category").Preload("Unit").
		Order("created_at ASC, id ASC").
		Find(&products).Error

	return products, err
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository creates a new category repository
func NewCategoryRepository(db *gorm.DB) domainRepo.CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *entity.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	var category entity.Category
	err := r.db.WithContext(ctx).First(&category, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &category, err
}

func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	var category entity.Category
	err := r.db.WithContext(ctx).First(&category, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &category, err
}

func (r *categoryRepository) Update(ctx context.Context, category *entity.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Category{}, "id = ?", id).Error
}

func (r *categoryRepository) List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Category, int64, error) {
	var categories []entity.Category
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Category{})
	if !skipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).
		Order("name ASC").
		Find(&categories).Error

	return categories, total, err
}

type unitRepository struct {
	db *gorm.DB
}

// NewUnitRepository creates a new unit repository
func NewUnitRepository(db *gorm.DB) domainRepo.UnitRepository {
	return &unitRepository{db: db}
}

func (r *unitRepository) Create(ctx context.Context, unit *entity.Unit) error {
	return r.db.WithContext(ctx).Create(unit).Error
}

func (r *unitRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Unit, error) {
	var unit entity.Unit
	err := r.db.WithContext(ctx).First(&unit, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &unit, err
}

func (r *unitRepository) GetBySlug(ctx context.Context, slug string) (*entity.Unit, error) {
	var unit entity.Unit
	err := r.db.WithContext(ctx).First(&unit, "slug = ?", slug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &unit, err
}

func (r *unitRepository) Update(ctx context.Context, unit *entity.Unit) error {
	return r.db.WithContext(ctx).Save(unit).Error
}

func (r *unitRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entity.Unit{}, "id = ?", id).Error
}

func (r *unitRepository) List(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, skipUserFilter bool) ([]entity.Unit, int64, error) {
	var units []entity.Unit
	var total int64

	query := r.db.WithContext(ctx).Model(&entity.Unit{})
	if !skipUserFilter {
		query = query.Where("user_id = ?", userID)
	}

	if search != "" {
		query = query.Where("name ILIKE ?", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	params.Validate()
	err := query.Offset(params.Offset()).Limit(params.PerPage).
		Order("name ASC").
		Find(&units).Error

	return units, total, err
}
