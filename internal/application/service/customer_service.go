package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// CustomerService handles customer-related operations
type CustomerService struct {
	customerRepo repository.CustomerRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(customerRepo repository.CustomerRepository) *CustomerService {
	return &CustomerService{customerRepo: customerRepo}
}

// CreateCustomerInput represents the create customer input
type CreateCustomerInput struct {
	UserID        uuid.UUID
	Name          string
	Email         *string
	Phone         *string
	KRAPin        *string
	Address       *string
	AccountHolder *string
	AccountNumber *string
	BankName      *string
}

// CreateCustomer creates a new customer
func (s *CustomerService) CreateCustomer(ctx context.Context, input *CreateCustomerInput) (*entity.Customer, error) {
	// Extract tenant ID from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	customer := &entity.Customer{
		TenantID:      tenantID,
		UserID:        input.UserID,
		Name:          input.Name,
		Email:         input.Email,
		Phone:         input.Phone,
		KRAPin:        input.KRAPin,
		Address:       input.Address,
		AccountHolder: input.AccountHolder,
		AccountNumber: input.AccountNumber,
		BankName:      input.BankName,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *CustomerService) GetCustomer(ctx context.Context, id uuid.UUID) (*entity.Customer, error) {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, apperror.NewNotFoundError("Customer")
	}
	return customer, nil
}

// ListCustomers lists customers. If isSuperAdmin is true, returns all customers.
func (s *CustomerService) ListCustomers(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, isSuperAdmin bool) (*pagination.PaginatedResult[entity.Customer], error) {
	customers, total, err := s.customerRepo.List(ctx, userID, params, search, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Page, params.PerPage, total)
	return pagination.NewPaginatedResult(customers, pag), nil
}

// ListCustomersWithCursor lists customers using cursor-based pagination. If isSuperAdmin is true, returns all customers.
func (s *CustomerService) ListCustomersWithCursor(ctx context.Context, userID uuid.UUID, params *pagination.CursorParams, search string, isSuperAdmin bool) (*pagination.CursorPaginatedResult[entity.Customer], error) {
	customers, err := s.customerRepo.ListWithCursor(ctx, userID, params, search, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	// Determine if there was a cursor provided (meaning we're not on first page)
	hasPrev := params.Cursor != ""

	// Build cursor pagination response
	cursorPag, items := pagination.NewCursorPagination(customers, params.Limit,
		func(c entity.Customer) string { return c.ID.String() },
		func(c entity.Customer) time.Time { return c.CreatedAt },
	)
	cursorPag.HasPrev = hasPrev

	return pagination.NewCursorPaginatedResult(items, cursorPag), nil
}

// UpdateCustomerInput represents the update customer input
type UpdateCustomerInput struct {
	UserID        uuid.UUID
	ID            uuid.UUID
	IsSuperAdmin  bool
	Name          *string
	Email         *string
	Phone         *string
	KRAPin        *string
	Address       *string
	AccountHolder *string
	AccountNumber *string
	BankName      *string
}

// UpdateCustomer updates a customer
func (s *CustomerService) UpdateCustomer(ctx context.Context, input *UpdateCustomerInput) (*entity.Customer, error) {
	customer, err := s.customerRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if customer == nil {
		return nil, apperror.NewNotFoundError("Customer")
	}

	// Super-admin can update any customer, regular users can only update their own
	if !input.IsSuperAdmin && customer.UserID != input.UserID {
		return nil, apperror.ErrForbidden
	}

	if input.Name != nil {
		customer.Name = *input.Name
	}
	if input.Email != nil {
		customer.Email = input.Email
	}
	if input.Phone != nil {
		customer.Phone = input.Phone
	}
	if input.KRAPin != nil {
		customer.KRAPin = input.KRAPin
	}
	if input.Address != nil {
		customer.Address = input.Address
	}
	if input.AccountHolder != nil {
		customer.AccountHolder = input.AccountHolder
	}
	if input.AccountNumber != nil {
		customer.AccountNumber = input.AccountNumber
	}
	if input.BankName != nil {
		customer.BankName = input.BankName
	}

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		return nil, err
	}

	return customer, nil
}

// DeleteCustomer deletes a customer
func (s *CustomerService) DeleteCustomer(ctx context.Context, userID, id uuid.UUID, isSuperAdmin bool) error {
	customer, err := s.customerRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if customer == nil {
		return apperror.NewNotFoundError("Customer")
	}

	// Super-admin can delete any customer, regular users can only delete their own
	if !isSuperAdmin && customer.UserID != userID {
		return apperror.ErrForbidden
	}

	return s.customerRepo.Delete(ctx, id)
}

// SupplierService handles supplier-related operations
type SupplierService struct {
	supplierRepo repository.SupplierRepository
}

// NewSupplierService creates a new supplier service
func NewSupplierService(supplierRepo repository.SupplierRepository) *SupplierService {
	return &SupplierService{supplierRepo: supplierRepo}
}

// CreateSupplierInput represents the create supplier input
type CreateSupplierInput struct {
	UserID        uuid.UUID
	Name          string
	Email         *string
	Phone         *string
	Address       *string
	ShopName      *string
	KRAPin        *string
	Type          string
	AccountHolder *string
	AccountNumber *string
	BankName      *string
}

// CreateSupplier creates a new supplier
func (s *SupplierService) CreateSupplier(ctx context.Context, input *CreateSupplierInput) (*entity.Supplier, error) {
	// Extract tenant ID from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	supplier := &entity.Supplier{
		TenantID:      tenantID,
		UserID:        input.UserID,
		Name:          input.Name,
		Email:         input.Email,
		Phone:         input.Phone,
		Address:       input.Address,
		ShopName:      input.ShopName,
		KRAPin:        input.KRAPin,
		AccountHolder: input.AccountHolder,
		AccountNumber: input.AccountNumber,
		BankName:      input.BankName,
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, err
	}

	return supplier, nil
}

// GetSupplier retrieves a supplier by ID
func (s *SupplierService) GetSupplier(ctx context.Context, id uuid.UUID) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, apperror.NewNotFoundError("Supplier")
	}
	return supplier, nil
}

// ListSuppliers lists suppliers. If isSuperAdmin is true, returns all suppliers.
func (s *SupplierService) ListSuppliers(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams, search string, isSuperAdmin bool) (*pagination.PaginatedResult[entity.Supplier], error) {
	suppliers, total, err := s.supplierRepo.List(ctx, userID, params, search, isSuperAdmin)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Page, params.PerPage, total)
	return pagination.NewPaginatedResult(suppliers, pag), nil
}

// UpdateSupplierInput represents the update supplier input
type UpdateSupplierInput struct {
	UserID        uuid.UUID
	ID            uuid.UUID
	IsSuperAdmin  bool
	Name          *string
	Email         *string
	Phone         *string
	Address       *string
	ShopName      *string
	KRAPin        *string
	Type          *string
	AccountHolder *string
	AccountNumber *string
	BankName      *string
}

// UpdateSupplier updates a supplier
func (s *SupplierService) UpdateSupplier(ctx context.Context, input *UpdateSupplierInput) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if supplier == nil {
		return nil, apperror.NewNotFoundError("Supplier")
	}

	// Super-admin can update any supplier, regular users can only update their own
	if !input.IsSuperAdmin && supplier.UserID != input.UserID {
		return nil, apperror.ErrForbidden
	}

	if input.Name != nil {
		supplier.Name = *input.Name
	}
	if input.Email != nil {
		supplier.Email = input.Email
	}
	if input.Phone != nil {
		supplier.Phone = input.Phone
	}
	if input.Address != nil {
		supplier.Address = input.Address
	}
	if input.ShopName != nil {
		supplier.ShopName = input.ShopName
	}
	if input.KRAPin != nil {
		supplier.KRAPin = input.KRAPin
	}
	if input.AccountHolder != nil {
		supplier.AccountHolder = input.AccountHolder
	}
	if input.AccountNumber != nil {
		supplier.AccountNumber = input.AccountNumber
	}
	if input.BankName != nil {
		supplier.BankName = input.BankName
	}

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, err
	}

	return supplier, nil
}

// DeleteSupplier deletes a supplier
func (s *SupplierService) DeleteSupplier(ctx context.Context, userID, id uuid.UUID, isSuperAdmin bool) error {
	supplier, err := s.supplierRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if supplier == nil {
		return apperror.NewNotFoundError("Supplier")
	}

	// Super-admin can delete any supplier, regular users can only delete their own
	if !isSuperAdmin && supplier.UserID != userID {
		return apperror.ErrForbidden
	}

	return s.supplierRepo.Delete(ctx, id)
}
