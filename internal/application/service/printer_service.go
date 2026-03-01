package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/domain/entity"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/printer"
)

// PrinterService handles receipt formatting and thermal printing.
type PrinterService struct {
	printer       printer.Printer
	orderRepo     repository.OrderRepository
	quotationRepo repository.QuotationRepository
	printerType   string
}

// NewPrinterService creates a new printer service.
func NewPrinterService(
	p printer.Printer,
	orderRepo repository.OrderRepository,
	quotationRepo repository.QuotationRepository,
	printerType string,
) *PrinterService {
	return &PrinterService{
		printer:       p,
		orderRepo:     orderRepo,
		quotationRepo: quotationRepo,
		printerType:   printerType,
	}
}

// PrinterStatus returns the current printer status information.
type PrinterStatus struct {
	Configured bool   `json:"configured"`
	Connected  bool   `json:"connected"`
	Type       string `json:"type"`
}

// GetStatus returns printer connection status.
func (s *PrinterService) GetStatus() *PrinterStatus {
	return &PrinterStatus{
		Configured: s.printerType != "none" && s.printerType != "",
		Connected:  s.printer.IsConnected(),
		Type:       s.printerType,
	}
}

// TestPrint sends a test page to the printer.
// Returns the receipt data so the handler can return it as JSON when printer is disabled.
func (s *PrinterService) TestPrint() (*entity.Receipt, error) {
	receipt := &entity.Receipt{
		Header: entity.ReceiptHeader{
			StoreName: "PRINTER TEST",
			Address:   "Test Address",
			Phone:     "+254 000 000 000",
		},
		InvoiceNo: "TEST-001",
		Date:      "Test Date",
		Cashier:   "System",
		Items: []entity.ReceiptItem{
			{Name: "Test Item 1", Quantity: 1, UnitPrice: 10.00, Total: 10.00},
			{Name: "Test Item 2", Quantity: 2, UnitPrice: 5.00, Total: 10.00},
		},
		SubTotal: 20.00,
		VAT:      0.00,
		Total:    20.00,
		Paid:     20.00,
		Due:      0.00,
	}

	data := FormatReceipt(receipt)
	if err := s.printer.Print(data); err != nil {
		return receipt, fmt.Errorf("test print failed: %w", err)
	}

	return receipt, nil
}

// PrintOrderReceipt fetches an order (with details) and prints its receipt.
func (s *PrinterService) PrintOrderReceipt(ctx context.Context, orderID uuid.UUID) (*entity.Receipt, error) {
	order, err := s.orderRepo.GetWithDetails(ctx, orderID)
	if err != nil {
		return nil, apperror.NewNotFoundError("Order")
	}

	receipt := &entity.Receipt{
		Header: entity.ReceiptHeader{
			StoreName: "Investify Store",
		},
		InvoiceNo:   order.InvoiceNo,
		Date:        order.OrderDate.Format("2006-01-02 15:04"),
		PaymentType: order.PaymentType,
		SubTotal:    float64(order.SubTotal) / 100,
		VAT:         float64(order.VAT) / 100,
		Total:       float64(order.Total) / 100,
		Paid:        float64(order.Pay) / 100,
		Due:         float64(order.Due) / 100,
	}

	if order.Customer != nil {
		receipt.Customer = order.Customer.Name
	}

	for _, d := range order.Details {
		item := entity.ReceiptItem{
			Quantity:  d.Quantity,
			UnitPrice: float64(d.UnitCost) / 100,
			Total:     float64(d.Total) / 100,
		}
		if d.Product.Name != "" {
			item.Name = d.Product.Name
		} else {
			item.Name = "Product"
		}
		receipt.Items = append(receipt.Items, item)
	}

	data := FormatReceipt(receipt)
	if err := s.printer.Print(data); err != nil {
		log.Printf("Printer error (order %s): %v", orderID, err)
		return receipt, fmt.Errorf("failed to print receipt: %w", err)
	}

	return receipt, nil
}

// PrintQuotationReceipt fetches a quotation (with details) and prints its receipt.
func (s *PrinterService) PrintQuotationReceipt(ctx context.Context, quotationID uuid.UUID) (*entity.Receipt, error) {
	quotation, err := s.quotationRepo.GetWithDetails(ctx, quotationID)
	if err != nil {
		return nil, apperror.NewNotFoundError("Quotation")
	}

	receipt := &entity.Receipt{
		Header: entity.ReceiptHeader{
			StoreName: "Investify Store",
		},
		InvoiceNo: quotation.Reference,
		Date:      quotation.Date.Format("2006-01-02 15:04"),
		SubTotal:  quotation.TotalAmount - quotation.TaxAmount,
		VAT:       quotation.TaxAmount,
		Total:     quotation.TotalAmount,
	}

	if quotation.Customer != nil {
		receipt.Customer = quotation.Customer.Name
	} else if quotation.CustomerName != "" {
		receipt.Customer = quotation.CustomerName
	}

	for _, d := range quotation.Details {
		item := entity.ReceiptItem{
			Name:      d.ProductName,
			Quantity:  d.Quantity,
			UnitPrice: d.UnitPrice,
			Total:     d.SubTotal,
		}
		if item.Name == "" {
			if d.Product.Name != "" {
				item.Name = d.Product.Name
			} else {
				item.Name = "Product"
			}
		}
		receipt.Items = append(receipt.Items, item)
	}

	data := FormatReceipt(receipt)
	if err := s.printer.Print(data); err != nil {
		log.Printf("Printer error (quotation %s): %v", quotationID, err)
		return receipt, fmt.Errorf("failed to print receipt: %w", err)
	}

	return receipt, nil
}

// FormatReceipt converts a Receipt into ESC/POS bytes.
func FormatReceipt(r *entity.Receipt) []byte {
	doc := printer.NewDocument(32) // 58mm paper = 32 chars

	// Header
	doc.SetAlign(printer.AlignCenter).
		SetBold(true).
		SetFontSize(printer.FontDouble).
		Text(r.Header.StoreName).
		SetFontSize(printer.FontNormal).
		SetBold(false)

	if r.Header.Address != "" {
		doc.Text(r.Header.Address)
	}
	if r.Header.Phone != "" {
		doc.Text(r.Header.Phone)
	}
	if r.Header.TaxID != "" {
		doc.TextF("Tax ID: %s", r.Header.TaxID)
	}

	doc.SetAlign(printer.AlignLeft).
		Separator('-')

	// Invoice info
	doc.KeyValue("Invoice:", r.InvoiceNo).
		KeyValue("Date:", r.Date)

	if r.Cashier != "" {
		doc.KeyValue("Cashier:", r.Cashier)
	}
	if r.Customer != "" {
		doc.KeyValue("Customer:", r.Customer)
	}
	if r.PaymentType != "" {
		doc.KeyValue("Payment:", r.PaymentType)
	}

	doc.Separator('-')

	// Items
	for _, item := range r.Items {
		doc.ItemLine(item.Quantity, item.Name, fmt.Sprintf("%.2f", item.Total))
		if item.Quantity > 1 {
			doc.TextF("  @ %.2f each", item.UnitPrice)
		}
	}

	doc.Separator('-')

	// Totals
	doc.KeyValue("Subtotal:", fmt.Sprintf("%.2f", r.SubTotal))
	if r.VAT > 0 {
		doc.KeyValue("VAT:", fmt.Sprintf("%.2f", r.VAT))
	}
	doc.SetBold(true).
		KeyValue("TOTAL:", fmt.Sprintf("%.2f", r.Total)).
		SetBold(false)

	if r.Paid > 0 {
		doc.KeyValue("Paid:", fmt.Sprintf("%.2f", r.Paid))
	}
	if r.Due > 0 {
		doc.KeyValue("Due:", fmt.Sprintf("%.2f", r.Due))
	}

	doc.Separator('-')

	// Footer
	doc.SetAlign(printer.AlignCenter).
		LineFeed().
		Text("Thank you for your business!").
		LineFeed().
		SetAlign(printer.AlignLeft)

	doc.FeedLines(3).
		PartialCut()

	return doc.Bytes()
}
