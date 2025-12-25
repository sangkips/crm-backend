package response

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sangkips/investify-api/pkg/apperror"
	"github.com/sangkips/investify-api/pkg/pagination"
)

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

// Meta contains metadata about the response
type Meta struct {
	Timestamp string `json:"timestamp"`
	RequestID string `json:"request_id"`
}

// newMeta creates metadata for the response
func newMeta(c *gin.Context) *Meta {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return &Meta{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		RequestID: requestID,
	}
}

// Success sends a success response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    newMeta(c),
	})
}

// SuccessWithPagination sends a success response with pagination
func SuccessWithPagination[T any](c *gin.Context, statusCode int, message string, result *pagination.PaginatedResult[T]) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    result,
		Meta:    newMeta(c),
	})
}

// Error sends an error response
func Error(c *gin.Context, err error) {
	appErr := apperror.GetAppError(err)
	c.JSON(appErr.Code, APIResponse{
		Success: false,
		Message: appErr.Message,
		Errors:  appErr.Errors,
		Meta:    newMeta(c),
	})
}

// ErrorWithCode sends an error response with a specific status code
func ErrorWithCode(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Message: message,
		Meta:    newMeta(c),
	})
}

// ValidationError sends a validation error response
func ValidationError(c *gin.Context, errors []apperror.FieldError) {
	c.JSON(422, APIResponse{
		Success: false,
		Message: "Validation failed",
		Errors:  errors,
		Meta:    newMeta(c),
	})
}

// Created sends a 201 Created response
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, 201, message, data)
}

// OK sends a 200 OK response
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, 200, message, data)
}

// NoContent sends a 204 No Content response
func NoContent(c *gin.Context) {
	c.Status(204)
}

// NotFound sends a 404 Not Found response
func NotFound(c *gin.Context, message string) {
	ErrorWithCode(c, 404, message)
}

// Unauthorized sends a 401 Unauthorized response
func Unauthorized(c *gin.Context, message string) {
	ErrorWithCode(c, 401, message)
}

// Forbidden sends a 403 Forbidden response
func Forbidden(c *gin.Context, message string) {
	ErrorWithCode(c, 403, message)
}

// BadRequest sends a 400 Bad Request response
func BadRequest(c *gin.Context, message string) {
	ErrorWithCode(c, 400, message)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, message string) {
	ErrorWithCode(c, 500, message)
}
