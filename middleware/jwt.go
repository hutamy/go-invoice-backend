package middleware

import (
	"net/http"
	"strings"

	"github.com/hutamy/go-invoice-backend/utils"
	"github.com/hutamy/go-invoice-backend/utils/errors"
	"github.com/labstack/echo/v4"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return utils.Response(c, http.StatusUnauthorized, errors.ErrInvalidToken.Error(), nil)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseJWT(tokenStr)
		if err != nil {
			return utils.Response(c, http.StatusUnauthorized, errors.ErrInvalidToken.Error(), nil)
		}

		c.Set("user_id", uint(claims["user_id"].(float64)))
		return next(c)
	}
}
