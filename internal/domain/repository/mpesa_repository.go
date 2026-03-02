package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
)

// MpesaTransactionRepository defines the interface for M-Pesa transaction data operations
type MpesaTransactionRepository interface {
	// Create creates a new M-Pesa transaction record
	Create(ctx context.Context, tx *entity.MpesaTransaction) error

	// GetByID retrieves a transaction by ID (tenant-scoped)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.MpesaTransaction, error)

	// GetByCheckoutRequestID retrieves a transaction by Daraja CheckoutRequestID (global — no tenant scope)
	GetByCheckoutRequestID(ctx context.Context, checkoutRequestID string) (*entity.MpesaTransaction, error)

	// GetByOrderID retrieves all transactions for an order (tenant-scoped)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]entity.MpesaTransaction, error)

	// Update updates an existing transaction
	Update(ctx context.Context, tx *entity.MpesaTransaction) error
}
