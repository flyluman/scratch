package httpapi

import "github.com/labstack/echo/v4"

type RouteRegistrar interface {
	RegisterRoutes(g *echo.Group)
}
