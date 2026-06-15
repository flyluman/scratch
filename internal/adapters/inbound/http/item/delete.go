package item

import (
	"net/http"

	"github.com/labstack/echo/v4"

	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
)

// deleteItem godoc
// @Summary     Delete an item
// @Description Soft or hard delete an item by ID (controlled by feature flag)
// @Tags        items
// @Accept      json
// @Produce     json
// @Param       id   path  string  true  "Item ID"
// @Success     200  {object}  httpapi.Envelope{data=object}
// @Failure     404  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Failure     401  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Router      /items/{id} [delete]
func (h *Handler) deleteItem(c echo.Context) error {
	if err := h.delete.Execute(c.Request().Context(), h.actorID(c), c.Param("id")); err != nil {
		return httpapi.WriteAppError(c, err)
	}
	return httpapi.WriteSuccess(c, http.StatusOK, map[string]string{"deleted": c.Param("id")})
}
