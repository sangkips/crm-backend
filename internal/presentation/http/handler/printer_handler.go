package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/request"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
)

// PrinterHandler handles printer-related HTTP requests.
type PrinterHandler struct {
	printerService *service.PrinterService
}

// NewPrinterHandler creates a new printer handler.
func NewPrinterHandler(printerService *service.PrinterService) *PrinterHandler {
	return &PrinterHandler{printerService: printerService}
}

// GetStatus returns the current printer connection status.
func (h *PrinterHandler) GetStatus(c *gin.Context) {
	status := h.printerService.GetStatus()
	response.OK(c, "Printer status retrieved", status)
}

// TestPrint sends a test page to the printer.
func (h *PrinterHandler) TestPrint(c *gin.Context) {
	receipt, err := h.printerService.TestPrint()
	if err != nil {
		// Return the receipt data anyway (useful when printer type is "none")
		response.OK(c, "Test print completed (printer may be disabled)", gin.H{
			"receipt": receipt,
			"warning": err.Error(),
		})
		return
	}

	response.OK(c, "Test page sent to printer", gin.H{
		"receipt": receipt,
	})
}

// PrintReceipt prints a receipt for an order or quotation.
func (h *PrinterHandler) PrintReceipt(c *gin.Context) {
	var req request.PrintReceiptRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		response.BadRequest(c, "Invalid ID format")
		return
	}

	ctx := c.Request.Context()

	switch req.Type {
	case "order":
		receipt, err := h.printerService.PrintOrderReceipt(ctx, id)
		if err != nil {
			// If receipt was built but printing failed, return receipt with warning
			if receipt != nil {
				response.OK(c, "Receipt generated but printing failed", gin.H{
					"receipt": receipt,
					"warning": err.Error(),
				})
				return
			}
			response.Error(c, err)
			return
		}
		response.OK(c, "Order receipt printed successfully", gin.H{
			"receipt": receipt,
		})

	case "quotation":
		receipt, err := h.printerService.PrintQuotationReceipt(ctx, id)
		if err != nil {
			if receipt != nil {
				response.OK(c, "Receipt generated but printing failed", gin.H{
					"receipt": receipt,
					"warning": err.Error(),
				})
				return
			}
			response.Error(c, err)
			return
		}
		response.OK(c, "Quotation receipt printed successfully", gin.H{
			"receipt": receipt,
		})

	default:
		response.ErrorWithCode(c, http.StatusBadRequest, "Invalid receipt type. Use 'order' or 'quotation'")
	}
}
