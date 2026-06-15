package item

import (
	httpapi "github.com/flyluman/scratch/internal/adapters/inbound/http"
	"github.com/flyluman/scratch/internal/application"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	create *application.CreateItemUseCase
	get    *application.GetItemUseCase
	list   *application.ListItemUseCase
	update *application.UpdateItemUseCase
	delete *application.DeleteItemUseCase
}

func NewHandler(mod *application.ItemModule) *Handler {
	return &Handler{
		create: mod.Create,
		get:    mod.Get,
		list:   mod.List,
		update: mod.Update,
		delete: mod.Delete,
	}
}

func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.POST("/items", h.createItem)
	g.GET("/items", h.listItems)
	g.GET("/items/:id", h.getItem)
	g.PUT("/items/:id", h.updateItem)
	g.DELETE("/items/:id", h.deleteItem)
}

func (h *Handler) actorID(c echo.Context) string {
	return httpapi.ActorID(c)
}
