package item

import (
	"net/http"

	"github.com/labstack/echo/v4"

	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
)

// getItem godoc
// @Summary     Get an item by ID
// @Description Retrieve a single item by its unique identifier
// @Tags        items
// @Accept      json
// @Produce     json
// @Param       id   path  string  true  "Item ID"
// @Success     200  {object}  httpapi.Envelope{data=ItemResponse}
// @Failure     404  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Failure     401  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Router      /items/{id} [get]
func (h *Handler) getItem(c echo.Context) error {
	item, err := h.get.Execute(c.Request().Context(), c.Param("id"))
	if err != nil {
		return httpapi.WriteAppError(c, err)
	}
	return httpapi.WriteSuccess(c, http.StatusOK, toItemResponse(item))
}
