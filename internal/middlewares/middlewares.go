package middlewares

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func JSONGuard(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Header.Get("Content-Type") != "application/json" {
			return echo.NewHTTPError(http.StatusBadRequest, "Available Content-Type is application/json")
		}

		return next(c)
	}
}
