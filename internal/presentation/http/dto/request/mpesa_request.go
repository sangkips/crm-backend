package request

import "github.com/google/uuid"

// STKPushRequest represents the request body for initiating an M-Pesa STK Push
type STKPushRequest struct {
	OrderID     uuid.UUID `json:"order_id" binding:"required"`
	PhoneNumber string    `json:"phone_number" binding:"required"`
}
