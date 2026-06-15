package item

import (
	"net/http"

	"github.com/flyluman/scratch/internal/application"
	"github.com/labstack/echo/v4"

	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
)

// CreateItemRequest represents the request body for creating an item.
type CreateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// createItem godoc
// @Summary     Create a new item
// @Description Create a new item with name and description
// @Tags        items
// @Accept      json
// @Produce     json
// @Param       request  body  CreateItemRequest  true  "Item to create"
// @Success     201  {object}  httpapi.Envelope{data=ItemResponse}
// @Failure     400  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Failure     401  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Router      /items [post]
func (h *Handler) createItem(c echo.Context) error {
	var req CreateItemRequest
	if err := c.Bind(&req); err != nil {
		return httpapi.WriteAppError(c, application.ErrValidation)
	}

	item, err := h.create.Execute(c.Request().Context(), h.actorID(c), req.Name, req.Description)
	if err != nil {
		return httpapi.WriteAppError(c, err)
	}

	return httpapi.WriteSuccess(c, http.StatusCreated, toItemResponse(item))
}
