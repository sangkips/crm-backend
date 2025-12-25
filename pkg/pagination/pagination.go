package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// =============================================================================
// Page-Based Pagination (Offset Pagination)
// =============================================================================

// Pagination represents pagination parameters
type Pagination struct {
	CurrentPage int   `json:"current_page"`
	PerPage     int   `json:"per_page"`
	Total       int64 `json:"total"`
	TotalPages  int   `json:"total_pages"`
	HasNext     bool  `json:"has_next"`
	HasPrev     bool  `json:"has_prev"`
}

// PaginationParams represents input parameters for pagination
type PaginationParams struct {
	Page    int `form:"page" json:"page"`
	PerPage int `form:"per_page" json:"per_page"`
}

// DefaultPagination returns default pagination values
func DefaultPagination() *PaginationParams {
	return &PaginationParams{
		Page:    1,
		PerPage: 15,
	}
}

// Validate ensures pagination parameters are within valid ranges
func (p *PaginationParams) Validate() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PerPage < 1 {
		p.PerPage = 15
	}
	if p.PerPage > 100 {
		p.PerPage = 100
	}
}

// Offset calculates the offset for SQL queries
func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// NewPagination creates a new Pagination response
func NewPagination(page, perPage int, total int64) *Pagination {
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))

	return &Pagination{
		CurrentPage: page,
		PerPage:     perPage,
		Total:       total,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrev:     page > 1,
	}
}

// PaginatedResult represents a paginated result with items and pagination info
type PaginatedResult[T any] struct {
	Items      []T         `json:"items"`
	Pagination *Pagination `json:"pagination"`
}

// NewPaginatedResult creates a new paginated result
func NewPaginatedResult[T any](items []T, pagination *Pagination) *PaginatedResult[T] {
	return &PaginatedResult[T]{
		Items:      items,
		Pagination: pagination,
	}
}

// =============================================================================
// Cursor-Based Pagination (Keyset Pagination)
// =============================================================================

// CursorDirection represents the direction of cursor navigation
type CursorDirection string

const (
	CursorDirectionNext CursorDirection = "next"
	CursorDirectionPrev CursorDirection = "prev"
)

// Cursor represents the decoded cursor data
type Cursor struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// CursorParams represents input parameters for cursor-based pagination
type CursorParams struct {
	Cursor    string          `form:"cursor" json:"cursor"`       // Base64 encoded cursor
	Direction CursorDirection `form:"direction" json:"direction"` // "next" or "prev"
	Limit     int             `form:"limit" json:"limit"`
}

// CursorPagination represents cursor-based pagination response metadata
type CursorPagination struct {
	NextCursor *string `json:"next_cursor,omitempty"` // Cursor to fetch next page
	PrevCursor *string `json:"prev_cursor,omitempty"` // Cursor to fetch previous page
	HasNext    bool    `json:"has_next"`
	HasPrev    bool    `json:"has_prev"`
	Limit      int     `json:"limit"`
}

// CursorPaginatedResult represents a cursor-paginated result with items
type CursorPaginatedResult[T any] struct {
	Items      []T               `json:"items"`
	Pagination *CursorPagination `json:"pagination"`
}

// DefaultCursorParams returns default cursor pagination values
func DefaultCursorParams() *CursorParams {
	return &CursorParams{
		Cursor:    "",
		Direction: CursorDirectionNext,
		Limit:     15,
	}
}

// Validate ensures cursor pagination parameters are within valid ranges
func (c *CursorParams) Validate() {
	if c.Limit < 1 {
		c.Limit = 15
	}
	if c.Limit > 100 {
		c.Limit = 100
	}
	if c.Direction == "" {
		c.Direction = CursorDirectionNext
	}
}

// DecodeCursor decodes a base64 cursor string into a Cursor struct
func (c *CursorParams) DecodeCursor() (*Cursor, error) {
	if c.Cursor == "" {
		return nil, nil
	}

	decoded, err := base64.URLEncoding.DecodeString(c.Cursor)
	if err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor data: %w", err)
	}

	return &cursor, nil
}

// EncodeCursor creates a base64 encoded cursor from an ID and optional timestamp
func EncodeCursor(id string, createdAt ...time.Time) string {
	cursor := Cursor{ID: id}
	if len(createdAt) > 0 {
		cursor.CreatedAt = createdAt[0]
	}

	data, _ := json.Marshal(cursor)
	return base64.URLEncoding.EncodeToString(data)
}

// NewCursorPagination creates a new CursorPagination response
// items should be the fetched items (with limit+1 to detect hasMore)
// limit is the requested limit
func NewCursorPagination[T any](items []T, limit int, getID func(T) string, getCreatedAt func(T) time.Time) (*CursorPagination, []T) {
	hasMore := len(items) > limit

	// Trim to the requested limit if we fetched extra
	if hasMore {
		items = items[:limit]
	}

	pagination := &CursorPagination{
		Limit:   limit,
		HasNext: hasMore,
		HasPrev: false, // Will be determined by the caller based on cursor presence
	}

	if len(items) > 0 {
		// Set next cursor from the last item
		lastItem := items[len(items)-1]
		nextCursor := EncodeCursor(getID(lastItem), getCreatedAt(lastItem))
		pagination.NextCursor = &nextCursor

		// Set prev cursor from the first item
		firstItem := items[0]
		prevCursor := EncodeCursor(getID(firstItem), getCreatedAt(firstItem))
		pagination.PrevCursor = &prevCursor
	}

	return pagination, items
}

// NewCursorPaginatedResult creates a new cursor-paginated result
func NewCursorPaginatedResult[T any](items []T, pagination *CursorPagination) *CursorPaginatedResult[T] {
	return &CursorPaginatedResult[T]{
		Items:      items,
		Pagination: pagination,
	}
}

// =============================================================================
// Unified Pagination (Supports Both Strategies)
// =============================================================================

// UnifiedPaginationParams accepts both page-based and cursor-based parameters
type UnifiedPaginationParams struct {
	// Page-based parameters
	Page    int `form:"page" json:"page"`
	PerPage int `form:"per_page" json:"per_page"`

	// Cursor-based parameters
	Cursor    string          `form:"cursor" json:"cursor"`
	Direction CursorDirection `form:"direction" json:"direction"`
	Limit     int             `form:"limit" json:"limit"`
}

// IsCursorBased returns true if cursor-based pagination is being used
func (u *UnifiedPaginationParams) IsCursorBased() bool {
	return u.Cursor != "" || u.Limit > 0
}

// ToPaginationParams converts to page-based params
func (u *UnifiedPaginationParams) ToPaginationParams() *PaginationParams {
	params := &PaginationParams{
		Page:    u.Page,
		PerPage: u.PerPage,
	}
	params.Validate()
	return params
}

// ToCursorParams converts to cursor-based params
func (u *UnifiedPaginationParams) ToCursorParams() *CursorParams {
	params := &CursorParams{
		Cursor:    u.Cursor,
		Direction: u.Direction,
		Limit:     u.Limit,
	}
	if params.Limit == 0 && u.PerPage > 0 {
		params.Limit = u.PerPage
	}
	params.Validate()
	return params
}

// UnifiedPaginatedResult represents a result that includes both pagination types
type UnifiedPaginatedResult[T any] struct {
	Items []T `json:"items"`

	// Page-based fields (present when using page-based pagination)
	CurrentPage *int   `json:"current_page,omitempty"`
	TotalPages  *int   `json:"total_pages,omitempty"`
	Total       *int64 `json:"total,omitempty"`

	// Cursor-based fields (present when using cursor-based pagination)
	NextCursor *string `json:"next_cursor,omitempty"`
	PrevCursor *string `json:"prev_cursor,omitempty"`

	// Common fields
	HasNext bool `json:"has_next"`
	HasPrev bool `json:"has_prev"`
	PerPage int  `json:"per_page"`
}

// NewUnifiedPaginatedResultFromPage creates a unified result from page-based pagination
func NewUnifiedPaginatedResultFromPage[T any](items []T, pagination *Pagination) *UnifiedPaginatedResult[T] {
	return &UnifiedPaginatedResult[T]{
		Items:       items,
		CurrentPage: &pagination.CurrentPage,
		TotalPages:  &pagination.TotalPages,
		Total:       &pagination.Total,
		HasNext:     pagination.HasNext,
		HasPrev:     pagination.HasPrev,
		PerPage:     pagination.PerPage,
	}
}

// NewUnifiedPaginatedResultFromCursor creates a unified result from cursor-based pagination
func NewUnifiedPaginatedResultFromCursor[T any](items []T, pagination *CursorPagination) *UnifiedPaginatedResult[T] {
	return &UnifiedPaginatedResult[T]{
		Items:      items,
		NextCursor: pagination.NextCursor,
		PrevCursor: pagination.PrevCursor,
		HasNext:    pagination.HasNext,
		HasPrev:    pagination.HasPrev,
		PerPage:    pagination.Limit,
	}
}
