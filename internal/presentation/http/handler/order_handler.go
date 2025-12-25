package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/internal/application/service"
	"github.com/sangkips/investify-api/internal/domain/enum"
	"github.com/sangkips/investify-api/internal/domain/repository"
	"github.com/sangkips/investify-api/internal/presentation/http/dto/response"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// List handles listing orders (supports both page-based and cursor-based pagination)
func (h *OrderHandler) List(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	// Check if cursor-based pagination is requested
	if cursor := c.Query("cursor"); cursor != "" || c.Query("limit") != "" {
		h.listWithCursor(c, *userID, isSuperAdmin)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))
	search := c.Query("search")
	statusStr := c.Query("status")

	params := &repository.OrderFilterParams{
		Pagination: &pagination.PaginationParams{
			Page:    page,
			PerPage: perPage,
		},
		Search:         search,
		SortBy:         c.Query("sort_by"),
		SortOrder:      c.Query("sort_order"),
		SkipUserFilter: isSuperAdmin,
	}

	if statusStr != "" {
		statusInt, err := strconv.Atoi(statusStr)
		if err == nil {
			status := enum.OrderStatus(statusInt)
			params.Status = &status
		}
	}

	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := uuid.Parse(customerIDStr); err == nil {
			params.CustomerID = &customerID
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			params.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			params.EndDate = &endDate
		}
	}

	result, err := h.orderService.ListOrders(c.Request.Context(), *userID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Orders retrieved successfully", result)
}

// listWithCursor handles listing orders with cursor-based pagination
func (h *OrderHandler) listWithCursor(c *gin.Context, userID uuid.UUID, isSuperAdmin bool) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "15"))
	cursor := c.Query("cursor")
	direction := c.DefaultQuery("direction", "next")
	search := c.Query("search")
	statusStr := c.Query("status")

	params := &repository.OrderCursorFilterParams{
		Cursor: &pagination.CursorParams{
			Cursor:    cursor,
			Direction: pagination.CursorDirection(direction),
			Limit:     limit,
		},
		Search:         search,
		SkipUserFilter: isSuperAdmin,
	}

	if statusStr != "" {
		statusInt, err := strconv.Atoi(statusStr)
		if err == nil {
			status := enum.OrderStatus(statusInt)
			params.Status = &status
		}
	}

	if customerIDStr := c.Query("customer_id"); customerIDStr != "" {
		if customerID, err := uuid.Parse(customerIDStr); err == nil {
			params.CustomerID = &customerID
		}
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			params.StartDate = &startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			params.EndDate = &endDate
		}
	}

	result, err := h.orderService.ListOrdersWithCursor(c.Request.Context(), userID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, 200, "Orders retrieved successfully", result)
}

// Create handles creating an order
func (h *OrderHandler) Create(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var req struct {
		CustomerID  *uuid.UUID `json:"customer_id"`
		PaymentType string     `json:"payment_type"`
		Pay         float64    `json:"pay"`
		Items       []struct {
			ProductID uuid.UUID `json:"product_id"`
			Quantity  int       `json:"quantity"`
			UnitCost  float64   `json:"unit_cost"`
		} `json:"items" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	items := make([]service.OrderItemInput, len(req.Items))
	for i, item := range req.Items {
		items[i] = service.OrderItemInput{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			UnitCost:  item.UnitCost,
		}
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), &service.CreateOrderInput{
		UserID:      *userID,
		CustomerID:  req.CustomerID,
		PaymentType: req.PaymentType,
		Pay:         req.Pay,
		Items:       items,
	})
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Created(c, "Order created successfully", order)
}

// Get handles getting a single order
func (h *OrderHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Order retrieved successfully", order)
}

// UpdateStatus handles updating order status
func (h *OrderHandler) UpdateStatus(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var req struct {
		Status int `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), *userID, id, enum.OrderStatus(req.Status)); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Order status updated successfully", nil)
}

// Cancel handles canceling an order
func (h *OrderHandler) Cancel(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	if err := h.orderService.CancelOrder(c.Request.Context(), *userID, id); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Order cancelled successfully", nil)
}

// GetDueOrders handles getting orders with dues
func (h *OrderHandler) GetDueOrders(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "15"))

	params := &pagination.PaginationParams{
		Page:    page,
		PerPage: perPage,
	}

	result, err := h.orderService.GetDueOrders(c.Request.Context(), *userID, params)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.SuccessWithPagination(c, 200, "Due orders retrieved successfully", result)
}

// PayDue handles paying a due amount
func (h *OrderHandler) PayDue(c *gin.Context) {
	userID := GetUserID(c)
	if userID == nil {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	isSuperAdmin := IsSuperAdmin(c)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "Invalid order ID")
		return
	}

	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.orderService.PayDue(c.Request.Context(), *userID, id, req.Amount, isSuperAdmin); err != nil {
		response.Error(c, err)
		return
	}

	response.OK(c, "Payment recorded successfully", nil)
}
