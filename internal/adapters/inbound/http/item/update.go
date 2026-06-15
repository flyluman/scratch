package item

import (
	"net/http"

	"github.com/flyluman/scratch/internal/application"
	"github.com/labstack/echo/v4"

	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
)

// UpdateItemRequest represents the request body for updating an item.
type UpdateItemRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// updateItem godoc
// @Summary     Update an existing item
// @Description Update the name and/or description of an existing item
// @Tags        items
// @Accept      json
// @Produce     json
// @Param       id       path   string             true  "Item ID"
// @Param       request  body   UpdateItemRequest  true  "Updated item data"
// @Success     200  {object}  httpapi.Envelope{data=ItemResponse}
// @Failure     400  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Failure     404  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Failure     401  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Router      /items/{id} [put]
func (h *Handler) updateItem(c echo.Context) error {
	var req UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return httpapi.WriteAppError(c, application.ErrValidation)
	}

	item, err := h.update.Execute(c.Request().Context(), h.actorID(c), c.Param("id"), req.Name, req.Description)
	if err != nil {
		return httpapi.WriteAppError(c, err)
	}

	return httpapi.WriteSuccess(c, http.StatusOK, toItemResponse(item))
}
