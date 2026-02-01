package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/pagination"
	"github.com/sangkips/investify-api/pkg/utils"
)

// CategoryService handles category-related operations
type CategoryService struct {
	categoryRepo repository.CategoryRepository
}

// NewCategoryService creates a new category service
func NewCategoryService(categoryRepo repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

// CreateCategoryInput represents the create category input
type CreateCategoryInput struct {
	UserID uuid.UUID
	Name   string
}

// CreateCategory creates a new category
func (s *CategoryService) CreateCategory(ctx context.Context, input *CreateCategoryInput) (*entity.Category, error) {
	// Extract tenant ID from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	slug := utils.Slugify(input.Name)

	// Check if slug already exists
	existing, err := s.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.NewConflictError("Category with this name already exists")
	}

	category := &entity.Category{
		TenantID: tenantID,
		UserID:   input.UserID,
		Name:     input.Name,
		Slug:     slug,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// GetCategory retrieves a category by ID
func (s *CategoryService) GetCategory(ctx context.Context, id uuid.UUID) (*entity.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, apperror.NewNotFoundError("Category")
	}
	return category, nil
}

// ListCategories lists categories. If isSuperAdmin is true, returns all categories.
func (s *CategoryService) ListCategories(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, isSuperAdmin bool) (*pagination.PaginatedResult[entity.Category], error) {
	categories, total, err := s.categoryRepo.List(ctx, userID, params, search, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Page, params.PerPage, total)
	return pagination.NewPaginatedResult(categories, pag), nil
}

// UpdateCategoryInput represents the update category input
type UpdateCategoryInput struct {
	UserID       uuid.UUID
	ID           uuid.UUID
	IsSuperAdmin bool
	Name         string
}

// UpdateCategory updates a category
func (s *CategoryService) UpdateCategory(ctx context.Context, input *UpdateCategoryInput) (*entity.Category, error) {
	category, err := s.categoryRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if category == nil {
		return nil, apperror.NewNotFoundError("Category")
	}

	// Super-admin can update any category, regular users can only update their own
	if !input.IsSuperAdmin && category.UserID != input.UserID {
		return nil, apperror.ErrForbidden
	}

	newSlug := utils.Slugify(input.Name)
	if newSlug != category.Slug {
		existing, err := s.categoryRepo.GetBySlug(ctx, newSlug)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != category.ID {
			return nil, apperror.NewConflictError("Category with this name already exists")
		}
		category.Slug = newSlug
	}

	category.Name = input.Name

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	return category, nil
}

// DeleteCategory deletes a category
func (s *CategoryService) DeleteCategory(ctx context.Context, userID, id uuid.UUID, isSuperAdmin bool) error {
	category, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if category == nil {
		return apperror.NewNotFoundError("Category")
	}

	// Super-admin can delete any category, regular users can only delete their own
	if !isSuperAdmin && category.UserID != userID {
		return apperror.ErrForbidden
	}

	return s.categoryRepo.Delete(ctx, id)
}

// UnitService handles unit-related operations
type UnitService struct {
	unitRepo repository.UnitRepository
}

// NewUnitService creates a new unit service
func NewUnitService(unitRepo repository.UnitRepository) *UnitService {
	return &UnitService{unitRepo: unitRepo}
}

// CreateUnitInput represents the create unit input
type CreateUnitInput struct {
	UserID    uuid.UUID
	Name      string
	ShortCode string
}

// CreateUnit creates a new unit
func (s *UnitService) CreateUnit(ctx context.Context, input *CreateUnitInput) (*entity.Unit, error) {
	// Extract tenant ID from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	slug := utils.Slugify(input.Name)

	existing, err := s.unitRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperror.NewConflictError("Unit with this name already exists")
	}

	unit := &entity.Unit{
		TenantID:  tenantID,
		UserID:    input.UserID,
		Name:      input.Name,
		Slug:      slug,
		ShortCode: input.ShortCode,
	}

	if err := s.unitRepo.Create(ctx, unit); err != nil {
		return nil, err
	}

	return unit, nil
}

// GetUnit retrieves a unit by ID
func (s *UnitService) GetUnit(ctx context.Context, id uuid.UUID) (*entity.Unit, error) {
	unit, err := s.unitRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if unit == nil {
		return nil, apperror.NewNotFoundError("Unit")
	}
	return unit, nil
}

// ListUnits lists units. If isSuperAdmin is true, returns all units.
func (s *UnitService) ListUnits(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, isSuperAdmin bool) (*pagination.PaginatedResult[entity.Unit], error) {
	units, total, err := s.unitRepo.List(ctx, userID, params, search, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Page, params.PerPage, total)
	return pagination.NewPaginatedResult(units, pag), nil
}

// UpdateUnitInput represents the update unit input
type UpdateUnitInput struct {
	UserID       uuid.UUID
	ID           uuid.UUID
	IsSuperAdmin bool
	Name         string
	ShortCode    string
}

// UpdateUnit updates a unit
func (s *UnitService) UpdateUnit(ctx context.Context, input *UpdateUnitInput) (*entity.Unit, error) {
	unit, err := s.unitRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if unit == nil {
		return nil, apperror.NewNotFoundError("Unit")
	}

	// Super-admin can update any unit, regular users can only update their own
	if !input.IsSuperAdmin && unit.UserID != input.UserID {
		return nil, apperror.ErrForbidden
	}

	newSlug := utils.Slugify(input.Name)
	if newSlug != unit.Slug {
		existing, err := s.unitRepo.GetBySlug(ctx, newSlug)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != unit.ID {
			return nil, apperror.NewConflictError("Unit with this name already exists")
		}
		unit.Slug = newSlug
	}

	unit.Name = input.Name
	unit.ShortCode = input.ShortCode

	if err := s.unitRepo.Update(ctx, unit); err != nil {
		return nil, err
	}

	return unit, nil
}

// DeleteUnit deletes a unit
func (s *UnitService) DeleteUnit(ctx context.Context, userID, id uuid.UUID, isSuperAdmin bool) error {
	unit, err := s.unitRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if unit == nil {
		return apperror.NewNotFoundError("Unit")
	}

	// Super-admin can delete any unit, regular users can only delete their own
	if !isSuperAdmin && unit.UserID != userID {
		return apperror.ErrForbidden
	}

	return s.unitRepo.Delete(ctx, id)
}
