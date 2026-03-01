package request

// PrintReceiptRequest is the request body for printing a receipt.
type PrintReceiptRequest struct {
	Type string `json:"type" binding:"required,oneof=order quotation"`
	ID   string `json:"id" binding:"required,uuid"`
}
