package idgen

import (
	"github.com/google/uuid"
)

// NewID generates a new UUID v4 string
func NewID() string {
	return uuid.New().String()
}
