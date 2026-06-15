package domain

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrInvalidItem = errors.New("invalid item")
)

type Item struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

func (i Item) Validate() error {
	if i.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidItem)
	}
	if len(i.Name) > 255 {
		return fmt.Errorf("%w: name exceeds 255 characters", ErrInvalidItem)
	}
	return nil
}
