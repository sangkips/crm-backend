package utils

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
)

// NewUUID generates a new UUID
func NewUUID() uuid.UUID {
	return uuid.New()
}

// ParseUUID parses a string into a UUID
func ParseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// Slugify converts a string to a URL-friendly slug
func Slugify(s string) string {
	// Convert to lowercase
	s = strings.ToLower(s)

	// Replace spaces with hyphens
	s = strings.ReplaceAll(s, " ", "-")

	// Remove non-alphanumeric characters except hyphens
	reg := regexp.MustCompile("[^a-z0-9-]")
	s = reg.ReplaceAllString(s, "")

	// Remove multiple consecutive hyphens
	reg = regexp.MustCompile("-+")
	s = reg.ReplaceAllString(s, "-")

	// Trim hyphens from start and end
	s = strings.Trim(s, "-")

	return s
}

// GenerateInvoiceNo generates a unique invoice number
func GenerateInvoiceNo(prefix string, number int) string {
	return prefix + strings.ToUpper(uuid.New().String()[:8])
}

// GenerateReferenceNo generates a unique reference number
func GenerateReferenceNo(prefix string, id int) string {
	return prefix + "-" + strings.ToUpper(uuid.New().String()[:8])
}

// GenerateProductCode generates a unique product code
func GenerateProductCode() string {
	return "PROD-" + strings.ToUpper(uuid.New().String()[:8])
}
