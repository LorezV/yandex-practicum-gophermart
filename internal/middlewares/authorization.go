package middlewares

import (
	"context"
	"fmt"
	"github.com/LorezV/go-diploma.git/internal/config"
	"github.com/LorezV/go-diploma.git/internal/repositories/userrepository"
	"github.com/LorezV/go-diploma.git/internal/utils"
	"github.com/dgrijalva/jwt-go/v4"
	"net/http"
	"strings"
)

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextUser := utils.ContextUser{IsValid: false}

		defer func() {
			r = r.WithContext(context.WithValue(r.Context(), utils.ContextKey("user"), contextUser))
			next.ServeHTTP(w, r)
		}()

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			return
		}
		if parts[0] != "Bearer" {
			return
		}

		token, err := jwt.ParseWithClaims(parts[1], &utils.Claims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(config.Config.SecretKey), nil
		})
		if err != nil {
			return
		}

		claims, ok := token.Claims.(*utils.Claims)
		if !ok || !token.Valid {
			return
		}

		user, err := userrepository.Get(r.Context(), "id", claims.UserID)
		if err != nil {
			return
		}

		contextUser.User = user
		contextUser.IsValid = true
	})
}
