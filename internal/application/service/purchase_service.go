package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// PurchaseService handles purchase-related operations
type PurchaseService struct {
	purchaseRepo       repository.PurchaseRepository
	purchaseDetailRepo repository.PurchaseDetailRepository
	productRepo        repository.ProductRepository
	supplierRepo       repository.SupplierRepository
}

// NewPurchaseService creates a new purchase service
func NewPurchaseService(
	purchaseRepo repository.PurchaseRepository,
	purchaseDetailRepo repository.PurchaseDetailRepository,
	productRepo repository.ProductRepository,
	supplierRepo repository.SupplierRepository,
) *PurchaseService {
	return &PurchaseService{
		purchaseRepo:       purchaseRepo,
		purchaseDetailRepo: purchaseDetailRepo,
		productRepo:        productRepo,
		supplierRepo:       supplierRepo,
	}
}

// PurchaseItemInput represents an item in a purchase
type PurchaseItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	UnitCost  float64
}

// CreatePurchaseInput represents the create purchase input
type CreatePurchaseInput struct {
	UserID        uuid.UUID
	SupplierID    *uuid.UUID
	TaxPercentage float64
	Items         []PurchaseItemInput
}

// CreatePurchase creates a new purchase with its details
func (s *PurchaseService) CreatePurchase(ctx context.Context, input *CreatePurchaseInput) (*entity.Purchase, error) {
	// Extract tenant ID from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	// Validate supplier if provided
	if input.SupplierID != nil {
		supplier, err := s.supplierRepo.GetByID(ctx, *input.SupplierID)
		if err != nil {
			return nil, err
		}
		if supplier == nil {
			return nil, apperror.NewNotFoundError("Supplier")
		}
	}

	// Batch fetch all products in one query (prevents N+1)
	productIDs := make([]uuid.UUID, len(input.Items))
	for i, item := range input.Items {
		productIDs[i] = item.ProductID
	}

	products, err := s.productRepo.GetByIDs(ctx, productIDs)
	if err != nil {
		return nil, err
	}

	// Create a map for quick lookup
	productMap := make(map[uuid.UUID]*entity.Product, len(products))
	for i := range products {
		productMap[products[i].ID] = &products[i]
	}

	// Calculate totals and validate products
	var totalAmount int64
	purchaseDetails := make([]entity.PurchaseDetail, 0, len(input.Items))

	for _, item := range input.Items {
		_, exists := productMap[item.ProductID]
		if !exists {
			return nil, apperror.NewNotFoundError(fmt.Sprintf("Product %s", item.ProductID))
		}

		unitCostCents := int64(item.UnitCost * 100)
		itemTotal := unitCostCents * int64(item.Quantity)
		totalAmount += itemTotal

		purchaseDetails = append(purchaseDetails, entity.PurchaseDetail{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  unitCostCents,
			Total:     itemTotal,
		})
	}

	// Calculate tax
	taxAmount := int64(float64(totalAmount) * input.TaxPercentage / 100)

	// Generate purchase number
	purchaseNo := fmt.Sprintf("PUR-%s", uuid.New().String()[:8])

	purchase := &entity.Purchase{
		TenantID:      tenantID,
		UserID:        input.UserID,
		SupplierID:    input.SupplierID,
		CreatedByID:   &input.UserID,
		Date:          time.Now(),
		PurchaseNo:    purchaseNo,
		Status:        enum.PurchaseStatusPending,
		TotalAmount:   float64(totalAmount+taxAmount) / 100, // Convert cents to float
		TaxPercentage: input.TaxPercentage,
		TaxAmount:     float64(taxAmount) / 100, // Convert cents to float
	}

	if err := s.purchaseRepo.Create(ctx, purchase); err != nil {
		return nil, err
	}

	// Create purchase details
	for i := range purchaseDetails {
		purchaseDetails[i].PurchaseID = purchase.ID
	}

	if err := s.purchaseDetailRepo.CreateBatch(ctx, purchaseDetails); err != nil {
		return nil, err
	}

	return s.purchaseRepo.GetWithDetails(ctx, purchase.ID)
}

// GetPurchase retrieves a purchase by ID
func (s *PurchaseService) GetPurchase(ctx context.Context, id uuid.UUID) (*entity.Purchase, error) {
	purchase, err := s.purchaseRepo.GetWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if purchase == nil {
		return nil, apperror.NewNotFoundError("Purchase")
	}
	return purchase, nil
}

// ListPurchases lists purchases with filtering
func (s *PurchaseService) ListPurchases(ctx context.Context, userID uuid.UUID, params *repository.PurchaseFilterParams) (*pagination.PaginatedResult[entity.Purchase], error) {
	purchases, total, err := s.purchaseRepo.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Pagination.Page, params.Pagination.PerPage, total)
	return pagination.NewPaginatedResult(purchases, pag), nil
}

// ApprovePurchase approves a purchase and updates stock
func (s *PurchaseService) ApprovePurchase(ctx context.Context, userID, purchaseID uuid.UUID, isSuperAdmin bool) error {
	purchase, err := s.purchaseRepo.GetWithDetails(ctx, purchaseID)
	if err != nil {
		return err
	}
	if purchase == nil {
		return apperror.NewNotFoundError("Purchase")
	}

	// Super-admin can approve any purchase, regular users can only approve their own
	if !isSuperAdmin && purchase.UserID != userID {
		return apperror.ErrForbidden
	}

	if purchase.Status == enum.PurchaseStatusApproved {
		return apperror.NewAppError(400, "Purchase is already approved")
	}

	// Build increment map for stock update
	stockIncrements := make(map[uuid.UUID]int)
	for _, detail := range purchase.Details {
		stockIncrements[detail.ProductID] = detail.Quantity
	}

	// Atomically add purchased quantities to stock
	if err := s.productRepo.AtomicIncrementBatch(ctx, stockIncrements); err != nil {
		return err
	}

	return s.purchaseRepo.UpdateStatus(ctx, purchaseID, enum.PurchaseStatusApproved, userID)
}

// DeletePurchase deletes a pending purchase
func (s *PurchaseService) DeletePurchase(ctx context.Context, userID, purchaseID uuid.UUID, isSuperAdmin bool) error {
	purchase, err := s.purchaseRepo.GetByID(ctx, purchaseID)
	if err != nil {
		return err
	}
	if purchase == nil {
		return apperror.NewNotFoundError("Purchase")
	}

	// Super-admin can delete any purchase, regular users can only delete their own
	if !isSuperAdmin && purchase.UserID != userID {
		return apperror.ErrForbidden
	}

	if purchase.Status == enum.PurchaseStatusApproved {
		return apperror.NewAppError(400, "Cannot delete an approved purchase")
	}

	// Delete details first
	if err := s.purchaseDetailRepo.DeleteByPurchaseID(ctx, purchaseID); err != nil {
		return err
	}

	return s.purchaseRepo.Delete(ctx, purchaseID)
}

// GetPendingPurchases returns pending purchases
func (s *PurchaseService) GetPendingPurchases(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) (*pagination.PaginatedResult[entity.Purchase], error) {
	purchases, total, err := s.purchaseRepo.GetPendingPurchases(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Page, params.PerPage, total)
	return pagination.NewPaginatedResult(purchases, pag), nil
}
