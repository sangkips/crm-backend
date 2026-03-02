package handler

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/request"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/pkg/mpesa"
)

// MpesaHandler handles M-Pesa payment HTTP requests
type MpesaHandler struct {
	mpesaService *service.MpesaService
}

// NewMpesaHandler creates a new M-Pesa handler
func NewMpesaHandler(mpesaService *service.MpesaService) *MpesaHandler {
	return &MpesaHandler{mpesaService: mpesaService}
}

// InitiateSTKPush handles POST /mpesa/stkpush
func (h *MpesaHandler) InitiateSTKPush(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req request.STKPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body. Required fields: order_id, phone_number")
		return
	}

	tx, err := h.mpesaService.InitiateSTKPush(c.Request.Context(), &service.STKPushInput{
		OrderID:     req.OrderID,
		PhoneNumber: req.PhoneNumber,
		UserID:      *userID,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "STK Push initiated. Check your phone for the M-Pesa PIN prompt.", tx)
}

// Callback handles POST /mpesa/callback — public endpoint for Safaricom
func (h *MpesaHandler) Callback(c *gin.Context) {
	var callbackBody mpesa.STKCallbackBody
	if err := c.ShouldBindJSON(&callbackBody); err != nil {
		log.Printf("M-Pesa callback: invalid JSON body: %v", err)
		// Return 200 to Safaricom even on parse error to prevent retries
		c.JSON(200, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
		return
	}

	if err := h.mpesaService.HandleCallback(c.Request.Context(), &callbackBody); err != nil {
		log.Printf("M-Pesa callback: processing error: %v", err)
	}

	// Always return 200 to Safaricom
	c.JSON(200, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
}

// GetTransaction handles GET /mpesa/transactions/:id
func (h *MpesaHandler) GetTransaction(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID")
		return
	}

	tx, err := h.mpesaService.GetTransaction(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Transaction retrieved successfully", tx)
}

// GetOrderTransactions handles GET /mpesa/transactions/order/:order_id
func (h *MpesaHandler) GetOrderTransactions(c *gin.Context) {
	orderID, err := uuid.Parse(c.Param("order_id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	transactions, err := h.mpesaService.GetOrderTransactions(c.Request.Context(), orderID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Transactions retrieved successfully", transactions)
}
