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

// OrderService handles order-related operations
type OrderService struct {
	orderRepo       repository.OrderRepository
	orderDetailRepo repository.OrderDetailRepository
	productRepo     repository.ProductRepository
	customerRepo    repository.CustomerRepository
}

// NewOrderService creates a new order service
func NewOrderService(
	orderRepo repository.OrderRepository,
	orderDetailRepo repository.OrderDetailRepository,
	productRepo repository.ProductRepository,
	customerRepo repository.CustomerRepository,
) *OrderService {
	return &OrderService{
		orderRepo:       orderRepo,
		orderDetailRepo: orderDetailRepo,
		productRepo:     productRepo,
		customerRepo:    customerRepo,
	}
}

// OrderItemInput represents an item in an order
type OrderItemInput struct {
	ProductID uuid.UUID
	Quantity  int
	UnitCost  float64
}

// CreateOrderInput represents the create order input
type CreateOrderInput struct {
	UserID      uuid.UUID
	CustomerID  *uuid.UUID
	PaymentType string
	Pay         float64
	Items       []OrderItemInput
}

// CreateOrder creates a new order with its details
func (s *OrderService) CreateOrder(ctx context.Context, input *CreateOrderInput) (*entity.Order, error) {
	// Extract tenant ID from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	// Validate customer if provided
	if input.CustomerID != nil {
		customer, err := s.customerRepo.GetByID(ctx, *input.CustomerID)
		if err != nil {
			return nil, err
		}
		if customer == nil {
			return nil, apperror.NewNotFoundError("Customer")
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

	// Validate all products exist and calculate totals
	var subTotal int64         // Total before any tax consideration
	var taxableAmount int64    // Amount from products with exclusive tax (need to add VAT)
	var nonTaxableAmount int64 // Amount from products with inclusive tax (VAT already included)
	var totalProducts int
	orderDetails := make([]entity.OrderDetail, 0, len(input.Items))
	stockDecrements := make(map[uuid.UUID]int)

	for _, item := range input.Items {
		product, exists := productMap[item.ProductID]
		if !exists {
			return nil, apperror.NewNotFoundError(fmt.Sprintf("Product %s", item.ProductID))
		}

		unitCostCents := int64(item.UnitCost * 100)
		itemTotal := unitCostCents * int64(item.Quantity)
		subTotal += itemTotal
		totalProducts += item.Quantity

		// Separate taxable and non-taxable amounts based on product's tax_type
		// TaxTypeExclusive (0) = VAT needs to be added
		// TaxTypeInclusive (1) = VAT already included in price
		if product.TaxType == enum.TaxTypeExclusive {
			taxableAmount += itemTotal
		} else {
			nonTaxableAmount += itemTotal
		}

		orderDetails = append(orderDetails, entity.OrderDetail{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  unitCostCents,
			Total:     itemTotal,
		})

		// Prepare atomic stock decrement
		stockDecrements[product.ID] = item.Quantity
	}

	// Atomically decrement stock - this is race-condition safe
	// If any product has insufficient stock, the entire operation fails
	failedIDs, err := s.productRepo.AtomicDecrementBatch(ctx, stockDecrements)
	if err != nil {
		return nil, err
	}

	// If any products failed due to insufficient stock
	if len(failedIDs) > 0 {
		// Build list of product names that failed
		var failedNames []string
		for _, id := range failedIDs {
			if product, exists := productMap[id]; exists {
				failedNames = append(failedNames, product.Name)
			}
		}
		return nil, apperror.NewAppError(400, fmt.Sprintf("Insufficient stock for: %v", failedNames))
	}

	// Calculate VAT (16% for Kenya)
	// For exclusive products: VAT is added on top
	// For inclusive products: VAT is already in price, extract it for display
	// additionalVat = taxableAmount * 0.16 (for exclusive, this is added to total)
	// includedVat = nonTaxableAmount * (0.16 / 1.16) (for inclusive, already in subtotal)
	additionalVat := int64(float64(taxableAmount) * 0.16)
	includedVat := int64(float64(nonTaxableAmount) * (0.16 / 1.16))

	// Total VAT shown = additional + included (for transparency to customer)
	vat := additionalVat + includedVat
	// Total = subTotal + only the additional VAT (included VAT is already in subTotal)
	total := subTotal + additionalVat
	payCents := int64(input.Pay * 100)
	due := total - payCents

	// Generate invoice number
	invoiceNo := fmt.Sprintf("INV-%s", uuid.New().String()[:8])

	order := &entity.Order{
		TenantID:      tenantID,
		UserID:        input.UserID,
		CustomerID:    input.CustomerID,
		OrderDate:     time.Now(),
		OrderStatus:   enum.OrderStatusPending,
		TotalProducts: totalProducts,
		SubTotal:      subTotal,
		VAT:           vat,
		Total:         total,
		InvoiceNo:     invoiceNo,
		PaymentType:   input.PaymentType,
		Pay:           payCents,
		Due:           due,
	}

	if due <= 0 {
		order.OrderStatus = enum.OrderStatusComplete
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		// Stock was already decremented - we need to restore it
		_ = s.productRepo.AtomicIncrementBatch(ctx, stockDecrements)
		return nil, err
	}

	// Set order ID on details
	for i := range orderDetails {
		orderDetails[i].OrderID = order.ID
	}

	if err := s.orderDetailRepo.CreateBatch(ctx, orderDetails); err != nil {
		// Restore stock on failure
		_ = s.productRepo.AtomicIncrementBatch(ctx, stockDecrements)
		return nil, err
	}

	return s.orderRepo.GetWithDetails(ctx, order.ID)
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(ctx context.Context, id uuid.UUID) (*entity.Order, error) {
	order, err := s.orderRepo.GetWithDetails(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.NewNotFoundError("Order")
	}
	return order, nil
}

// ListOrders lists orders with filtering
func (s *OrderService) ListOrders(ctx context.Context, userID uuid.UUID, params *repository.OrderFilterParams) (*pagination.PaginatedResult[entity.Order], error) {
	orders, total, err := s.orderRepo.List(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Pagination.Page, params.Pagination.PerPage, total)
	return pagination.NewPaginatedResult(orders, pag), nil
}

// ListOrdersWithCursor lists orders with cursor-based pagination
func (s *OrderService) ListOrdersWithCursor(ctx context.Context, userID uuid.UUID, params *repository.OrderCursorFilterParams) (*pagination.CursorPaginatedResult[entity.Order], error) {
	orders, err := s.orderRepo.ListWithCursor(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	hasPrev := params.Cursor.Cursor != ""

	cursorPag, items := pagination.NewCursorPagination(orders, params.Cursor.Limit,
		func(o entity.Order) string { return o.ID.String() },
		func(o entity.Order) time.Time { return o.CreatedAt },
	)
	cursorPag.HasPrev = hasPrev

	return pagination.NewCursorPaginatedResult(items, cursorPag), nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, userID, orderID uuid.UUID, status enum.OrderStatus) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return apperror.NewNotFoundError("Order")
	}

	if order.UserID != userID {
		return apperror.ErrForbidden
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, status)
}

// CancelOrder cancels an order and restores stock
func (s *OrderService) CancelOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	order, err := s.orderRepo.GetWithDetails(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return apperror.NewNotFoundError("Order")
	}

	if order.UserID != userID {
		return apperror.ErrForbidden
	}

	if order.OrderStatus == enum.OrderStatusCancel {
		return apperror.NewAppError(400, "Order is already cancelled")
	}

	// Build increment map for stock restoration
	stockIncrements := make(map[uuid.UUID]int)
	for _, detail := range order.Details {
		stockIncrements[detail.ProductID] = detail.Quantity
	}

	// Atomically restore stock
	if err := s.productRepo.AtomicIncrementBatch(ctx, stockIncrements); err != nil {
		return err
	}

	return s.orderRepo.UpdateStatus(ctx, orderID, enum.OrderStatusCancel)
}

// GetDueOrders returns orders with outstanding dues
func (s *OrderService) GetDueOrders(ctx context.Context, userID uuid.UUID, params *pagination.PaginationParams) (*pagination.PaginatedResult[entity.Order], error) {
	orders, total, err := s.orderRepo.GetDueOrders(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	pag := pagination.NewPagination(params.Page, params.PerPage, total)
	return pagination.NewPaginatedResult(orders, pag), nil
}

// PayDue records a payment towards an order's due amount
func (s *OrderService) PayDue(ctx context.Context, userID, orderID uuid.UUID, amount float64, skipUserCheck bool) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return apperror.NewNotFoundError("Order")
	}

	// Only check ownership if not skipping (i.e., non-super-admin)
	if !skipUserCheck && order.UserID != userID {
		return apperror.ErrForbidden
	}

	amountCents := int64(amount * 100)
	order.Pay += amountCents
	order.Due -= amountCents

	if order.Due <= 0 {
		order.Due = 0
		order.OrderStatus = enum.OrderStatusComplete
	}

	return s.orderRepo.Update(ctx, order)
}
