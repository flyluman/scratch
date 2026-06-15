package item

import (
	"time"

	"github.com/flyluman/scratch/internal/domain"
)

// ItemResponse represents a single item in API responses.
type ItemResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// ListResponse represents a paginated list of items.
type ListResponse struct {
	Items      []ItemResponse `json:"items"`
	NextCursor string         `json:"next_cursor"`
}

func toItemResponse(item domain.Item) ItemResponse {
	return ItemResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
	}
}
