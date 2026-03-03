package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	infraRepo "github.com/sangkips/investify-api/internal/infrastructure/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/mpesa"
)

// MpesaService handles M-Pesa payment operations
type MpesaService struct {
	mpesaTxRepo  repository.MpesaTransactionRepository
	tenantRepo   repository.TenantRepository
	orderRepo    repository.OrderRepository
	orderService *OrderService
}

// NewMpesaService creates a new MpesaService
func NewMpesaService(
	mpesaTxRepo repository.MpesaTransactionRepository,
	tenantRepo repository.TenantRepository,
	orderRepo repository.OrderRepository,
	orderService *OrderService,
) *MpesaService {
	return &MpesaService{
		mpesaTxRepo:  mpesaTxRepo,
		tenantRepo:   tenantRepo,
		orderRepo:    orderRepo,
		orderService: orderService,
	}
}

// STKPushInput represents the input for initiating an STK Push
type STKPushInput struct {
	OrderID     uuid.UUID
	PhoneNumber string
	UserID      uuid.UUID
}

// InitiateSTKPush starts an M-Pesa STK Push payment for an order
func (s *MpesaService) InitiateSTKPush(ctx context.Context, input *STKPushInput) (*entity.MpesaTransaction, error) {
	// 1. Extract tenant from context
	tenantID, ok := infraRepo.GetTenantID(ctx)
	if !ok {
		return nil, apperror.NewBadRequestError("Tenant context required")
	}

	// 2. Fetch tenant and extract M-Pesa config
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tenant: %w", err)
	}
	if tenant == nil {
		return nil, apperror.NewNotFoundError("Tenant")
	}

	mpesaCfg := tenant.Settings.Mpesa
	if mpesaCfg == nil {
		return nil, apperror.NewBadRequestError("M-Pesa integration is not configured for this tenant. Please configure it in tenant settings.")
	}
	if mpesaCfg.ConsumerKey == "" || mpesaCfg.ConsumerSecret == "" || mpesaCfg.ShortCode == "" || mpesaCfg.PassKey == "" {
		return nil, apperror.NewBadRequestError("M-Pesa credentials are incomplete. Please update your tenant settings with all required fields.")
	}

	// 3. Fetch order and validate
	order, err := s.orderRepo.GetByID(ctx, input.OrderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, apperror.NewNotFoundError("Order")
	}
	if order.Due <= 0 {
		return nil, apperror.NewBadRequestError("Order has no outstanding balance")
	}
	if order.OrderStatus == enum.OrderStatusCancel {
		return nil, apperror.NewBadRequestError("Cannot pay for a cancelled order")
	}

	// 4. Create Daraja client for tenant's environment
	client := mpesa.NewClient(mpesaCfg.Environment)

	// 5. Get access token
	accessToken, err := client.GetAccessToken(mpesaCfg.ConsumerKey, mpesaCfg.ConsumerSecret)
	if err != nil {
		return nil, apperror.NewAppError(502, fmt.Sprintf("Failed to authenticate with M-Pesa: %v", err))
	}

	// 6. Build STK Push request
	timestamp := mpesa.GenerateTimestamp()
	password := mpesa.GeneratePassword(mpesaCfg.ShortCode, mpesaCfg.PassKey, timestamp)

	// Amount in whole KES (M-Pesa doesn't support cents)
	amount := int(order.Due / 100)
	if amount < 1 {
		amount = 1 // Minimum 1 KES
	}

	callbackURL := mpesaCfg.CallbackURL
	if callbackURL == "" {
		return nil, apperror.NewBadRequestError("M-Pesa callback URL is not configured. Please update your tenant settings.")
	}

	stkReq := mpesa.STKPushRequest{
		BusinessShortCode: mpesaCfg.ShortCode,
		Password:          password,
		Timestamp:         timestamp,
		TransactionType:   mpesa.TransactionTypeCustomerPayBillOnline,
		Amount:            amount,
		PartyA:            input.PhoneNumber,
		PartyB:            mpesaCfg.ShortCode,
		PhoneNumber:       input.PhoneNumber,
		CallBackURL:       callbackURL,
		AccountReference:  order.InvoiceNo,
		TransactionDesc:   fmt.Sprintf("Payment for %s", order.InvoiceNo),
	}

	// 7. Send STK Push
	stkResp, err := client.STKPush(accessToken, stkReq)
	if err != nil {
		return nil, apperror.NewAppError(502, fmt.Sprintf("M-Pesa STK Push failed: %v", err))
	}

	// 8. Save transaction record
	amountCents := int64(amount) * 100
	tx := &entity.MpesaTransaction{
		TenantID:          tenantID,
		OrderID:           input.OrderID,
		PhoneNumber:       input.PhoneNumber,
		Amount:            amountCents,
		MerchantRequestID: stkResp.MerchantRequestID,
		CheckoutRequestID: stkResp.CheckoutRequestID,
		Status:            enum.MpesaStatusPending,
	}

	if err := s.mpesaTxRepo.Create(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}

	return tx, nil
}

// HandleCallback processes the M-Pesa callback from Safaricom
func (s *MpesaService) HandleCallback(ctx context.Context, callbackBody *mpesa.STKCallbackBody) error {
	cb := callbackBody.Body.StkCallback

	// 1. Look up transaction by CheckoutRequestID (global scope — no tenant filter)
	tx, err := s.mpesaTxRepo.GetByCheckoutRequestID(ctx, cb.CheckoutRequestID)
	if err != nil {
		return fmt.Errorf("failed to look up transaction: %w", err)
	}
	if tx == nil {
		log.Printf("M-Pesa callback: no transaction found for CheckoutRequestID=%s", cb.CheckoutRequestID)
		return nil // Don't return error — Safaricom will retry
	}

	// 2. Update transaction with callback result
	tx.ResultCode = &cb.ResultCode
	tx.ResultDesc = cb.ResultDesc
	tx.MpesaReceiptNumber = cb.GetMpesaReceiptNumber()

	if cb.ResultCode == 0 {
		tx.Status = enum.MpesaStatusSuccess
	} else if cb.ResultCode == 1032 {
		// 1032 = Request cancelled by user
		tx.Status = enum.MpesaStatusCancelled
	} else {
		tx.Status = enum.MpesaStatusFailed
	}

	if err := s.mpesaTxRepo.Update(ctx, tx); err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}

	// 3. If payment was successful, update the order
	if tx.Status == enum.MpesaStatusSuccess {
		// Create a context with the transaction's tenant so order service works correctly
		orderCtx := infraRepo.WithTenant(ctx, tx.TenantID)
		amountKES := float64(tx.Amount) / 100.0

		if err := s.orderService.PayDue(orderCtx, uuid.Nil, tx.OrderID, amountKES, true); err != nil {
			log.Printf("M-Pesa callback: failed to update order payment for OrderID=%s: %v", tx.OrderID, err)
			// Don't return error — the transaction itself was recorded successfully
		}
	}

	return nil
}

// GetTransaction retrieves an M-Pesa transaction by ID (tenant-scoped)
func (s *MpesaService) GetTransaction(ctx context.Context, id uuid.UUID) (*entity.MpesaTransaction, error) {
	tx, err := s.mpesaTxRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		return nil, apperror.NewNotFoundError("M-Pesa transaction")
	}
	return tx, nil
}

// GetOrderTransactions retrieves all M-Pesa transactions for an order (tenant-scoped)
func (s *MpesaService) GetOrderTransactions(ctx context.Context, orderID uuid.UUID) ([]entity.MpesaTransaction, error) {
	return s.mpesaTxRepo.GetByOrderID(ctx, orderID)
}
