package middlewares

import (
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/services"
	"github.com/golang-jwt/jwt"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type Authorization struct {
	Services *services.Services
}

func (middleware *Authorization) Handle(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authorization := c.Request().Header.Get("Authorization")
		if authorization == "" {
			return echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("authorization is required"))
		}

		tokenString := strings.Split(authorization, " ")[1]

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(middleware.Services.Auth.GetSecret()), nil
		})

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			user, err := middleware.Services.User.FindByLogin(c.Request().Context(), claims["sub"].(string))
			if err != nil || user == nil {
				log.Error().Err(err).Msg("Finding user after valid JWT")
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			c.Set("user", user)
		} else {
			log.Error().Err(err).Msg("Parse JWT")
			return echo.NewHTTPError(http.StatusUnauthorized)
		}

		return next(c)
	}
}
