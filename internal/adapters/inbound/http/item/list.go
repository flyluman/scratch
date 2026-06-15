package item

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
)

// listItems godoc
// @Summary     List items with cursor pagination
// @Description Retrieve a paginated list of items
// @Tags        items
// @Accept      json
// @Produce     json
// @Param       cursor  query  string  false  "Cursor for pagination"
// @Param       limit   query  int     false  "Max items per page (1-100)"
// @Success     200  {object}  httpapi.Envelope{data=ListResponse}
// @Failure     401  {object}  httpapi.Envelope{error=httpapi.ErrBody}
// @Router      /items [get]
func (h *Handler) listItems(c echo.Context) error {
	limit := 20
	if v := c.QueryParam("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			limit = parsed
		}
	}

	items, nextCursor, err := h.list.Execute(c.Request().Context(), c.QueryParam("cursor"), limit)
	if err != nil {
		return httpapi.WriteAppError(c, err)
	}

	responses := make([]ItemResponse, len(items))
	for i, item := range items {
		responses[i] = toItemResponse(item)
	}

	return httpapi.WriteSuccess(c, http.StatusOK, ListResponse{Items: responses, NextCursor: nextCursor})
}
