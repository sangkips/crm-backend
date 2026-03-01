package entity

// ReceiptHeader holds the store/business header printed at the top of a receipt.
type ReceiptHeader struct {
	StoreName string `json:"store_name"`
	Address   string `json:"address,omitempty"`
	Phone     string `json:"phone,omitempty"`
	TaxID     string `json:"tax_id,omitempty"`
}

// ReceiptItem represents a single line item on a receipt.
type ReceiptItem struct {
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Total     float64 `json:"total"`
}

// Receipt is a value object representing a printable receipt.
// It is NOT a database entity — it is composed from order/quotation data at print time.
type Receipt struct {
	Header      ReceiptHeader `json:"header"`
	InvoiceNo   string        `json:"invoice_no"`
	Date        string        `json:"date"`
	Cashier     string        `json:"cashier,omitempty"`
	Customer    string        `json:"customer,omitempty"`
	PaymentType string        `json:"payment_type,omitempty"`
	Items       []ReceiptItem `json:"items"`
	SubTotal    float64       `json:"sub_total"`
	VAT         float64       `json:"vat"`
	Total       float64       `json:"total"`
	Paid        float64       `json:"paid"`
	Due         float64       `json:"due"`
}
