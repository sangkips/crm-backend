package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// QuotationService handles quotation-related operations
type QuotationService struct {
	quotationRepo       repository.QuotationRepository
	quotationDetailRepo repository.QuotationDetailRepository
	productRepo         repository.ProductRepository
	customerRepo        repository.CustomerRepository
}

// NewQuotationService creates a new quotation service
func NewQuotationService(
	quotationRepo repository.QuotationRepository,
	quotationDetailRepo repository.QuotationDetailRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
) *QuotationService {
	return &QuotationService{
		quotationRepo:       quotationRepo,
		quotationDetailRepo: quotationDetailRepo,
		productRepo:         productRepo,
		customerRepo:        customerRepo,
	}
}

// CreateQuotationInput represents the input for creating a quotation
type CreateQuotationInput struct {
	UserID             uuid.UUID
	CustomerID         *uuid.UUID
	Date               time.Time
	TaxPercentage      float64
	DiscountPercentage float64
	ShippingAmount     float64
	Note               *string
	Status             enum.QuotationStatus
	Items              []QuotationItemInput
}

// QuotationItemInput represents a line item input
type QuotationItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	UnitPrice float64
}

// CreateQuotation creates a new quotation
func (s *QuotationService) CreateQuotation(ctx context.Context, input *CreateQuotationInput) (*entity.Quotation, error) {
	// Generate reference number
	nextNum, err := s.quotationRepo.GetNextReferenceNumber(ctx)
	if err != nil {
		return nil, err
	}
	reference := fmt.Sprintf("QT-%06d", nextNum)

	// Get customer name if customer ID is provided
	var customerName string
	if input.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *input.CustomerID)
		if err != nil {
			return nil, err
		}
		if customer != nil {
			customerName = customer.Name
		}
	}

	// Calculate subtotal
	var subtotal float64
	for _, item := range input.Items {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}

	// Calculate tax and discount amounts
	taxAmount := (subtotal * input.TaxPercentage) / 100
	discountAmount := (subtotal * input.DiscountPercentage) / 100
	totalAmount := subtotal + taxAmount - discountAmount + input.ShippingAmount

	quotation := &entity.Quotation{
		UserID:             input.UserID,
		CustomerID:         input.CustomerID,
		Date:               input.Date,
		Reference:          reference,
		CustomerName:       customerName,
		TaxPercentage:      input.TaxPercentage,
		TaxAmount:          taxAmount,
		DiscountPercentage: input.DiscountPercentage,
		DiscountAmount:     discountAmount,
		ShippingAmount:     input.ShippingAmount,
		TotalAmount:        totalAmount,
		Status:             input.Status,
		Note:               input.Note,
	}

	if err := s.quotationRepo.Create(ctx, quotation); err != nil {
		return nil, err
	}

	// Create quotation details
	for _, item := range input.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, apperror.NewNotFoundError("Product")
		}

		detail := &entity.QuotationDetail{
			QuotationID: quotation.ID,
			ProductID:   item.ProductID,
			ProductName: product.Name,
			ProductCode: product.Code,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			SubTotal:    float64(item.Quantity) * item.UnitPrice,
		}

		if err := s.quotationDetailRepo.Create(ctx, detail); err != nil {
			return nil, err
		}
	}

	// Fetch the complete quotation with details
	return s.quotationRepo.GetWithDetails(ctx, quotation.ID)
}

// GetQuotation retrieves a quotation by ID
func (s *QuotationService) GetQuotation(ctx context.Context, id uuid.UUID) (*entity.Quotation, error) {
	quotation, err := s.quotationRepo.GetWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if quotation == nil {
		return nil, apperror.NewNotFoundError("Quotation")
	}
	return quotation, nil
}

// ListQuotationsInput represents the input for listing quotations
type ListQuotationsInput struct {
	UserID       uuid.UUID
	IsSuperAdmin bool
	Pagination   *pagination.PaginationParams
	Search       string
	Status       *enum.QuotationStatus
	CustomerID   *uuid.UUID
}

// ListQuotations lists quotations with filtering
func (s *QuotationService) ListQuotations(ctx context.Context, input *ListQuotationsInput) (*pagination.PaginatedResult[entity.Quotation], error) {
	params := &repository.QuotationFilterParams{
		Pagination: input.Pagination,
		Search:     input.Search,
		Status:     input.Status,
		CustomerID: input.CustomerID,
	}

	var userID uuid.UUID
	if !input.IsSuperAdmin {
		userID = input.UserID
	}

	quotations, total, err := s.quotationRepo.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(input.Pagination.Page, input.Pagination.PerPage, total)
	return pagination.NewPaginatedResult(quotations, pag), nil
}

// UpdateQuotationInput represents the input for updating a quotation
type UpdateQuotationInput struct {
	UserID             uuid.UUID
	ID                 uuid.UUID
	IsSuperAdmin       bool
	CustomerID         *uuid.UUID
	Date               time.Time
	TaxPercentage      float64
	DiscountPercentage float64
	ShippingAmount     float64
	Note               *string
	Status             enum.QuotationStatus
	Items              []QuotationItemInput
}

// UpdateQuotation updates an existing quotation
func (s *QuotationService) UpdateQuotation(ctx context.Context, input *UpdateQuotationInput) (*entity.Quotation, error) {
	quotation, err := s.quotationRepo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if quotation == nil {
		return nil, apperror.NewNotFoundError("Quotation")
	}

	// Check permission
	if !input.IsSuperAdmin && quotation.UserID != input.UserID {
		return nil, apperror.ErrForbidden
	}

	// Get customer name if customer ID is provided
	var customerName string
	if input.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *input.CustomerID)
		if err != nil {
			return nil, err
		}
		if customer != nil {
			customerName = customer.Name
		}
	}

	// Calculate subtotal
	var subtotal float64
	for _, item := range input.Items {
		subtotal += float64(item.Quantity) * item.UnitPrice
	}

	// Calculate tax and discount amounts
	taxAmount := (subtotal * input.TaxPercentage) / 100
	discountAmount := (subtotal * input.DiscountPercentage) / 100
	totalAmount := subtotal + taxAmount - discountAmount + input.ShippingAmount

	// Update quotation fields
	quotation.CustomerID = input.CustomerID
	quotation.Date = input.Date
	quotation.CustomerName = customerName
	quotation.TaxPercentage = input.TaxPercentage
	quotation.TaxAmount = taxAmount
	quotation.DiscountPercentage = input.DiscountPercentage
	quotation.DiscountAmount = discountAmount
	quotation.ShippingAmount = input.ShippingAmount
	quotation.TotalAmount = totalAmount
	quotation.Status = input.Status
	quotation.Note = input.Note

	if err := s.quotationRepo.Update(ctx, quotation); err != nil {
		return nil, err
	}

	// Delete existing details and create new ones
	if err := s.quotationDetailRepo.DeleteByQuotationID(ctx, quotation.ID); err != nil {
		return nil, err
	}

	for _, item := range input.Items {
		product, err := s.productRepo.GetByID(ctx, item.ProductID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return nil, apperror.NewNotFoundError("Product")
		}

		detail := &entity.QuotationDetail{
			QuotationID: quotation.ID,
			ProductID:   item.ProductID,
			ProductName: product.Name,
			ProductCode: product.Code,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			SubTotal:    float64(item.Quantity) * item.UnitPrice,
		}

		if err := s.quotationDetailRepo.Create(ctx, detail); err != nil {
			return nil, err
		}
	}

	return s.quotationRepo.GetWithDetails(ctx, quotation.ID)
}

// DeleteQuotation deletes a quotation
func (s *QuotationService) DeleteQuotation(ctx context.Context, userID, id uuid.UUID, isSuperAdmin bool) error {
	quotation, err := s.quotationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if quotation == nil {
		return apperror.NewNotFoundError("Quotation")
	}

	// Check permission
	if !isSuperAdmin && quotation.UserID != userID {
		return apperror.ErrForbidden
	}

	// Delete details first
	if err := s.quotationDetailRepo.DeleteByQuotationID(ctx, id); err != nil {
		return err
	}

	return s.quotationRepo.Delete(ctx, id)
}

// UpdateQuotationStatus updates the status of a quotation
func (s *QuotationService) UpdateQuotationStatus(ctx context.Context, userID, id uuid.UUID, status enum.QuotationStatus, isSuperAdmin bool) error {
	quotation, err := s.quotationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if quotation == nil {
		return apperror.NewNotFoundError("Quotation")
	}

	// Check permission
	if !isSuperAdmin && quotation.UserID != userID {
		return apperror.ErrForbidden
	}

	return s.quotationRepo.UpdateStatus(ctx, id, status)
}
