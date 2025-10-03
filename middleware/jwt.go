package middleware

import (
    "net/http"
    "strings"

    httpresp "github.com/hutamy/go-invoice-backend/internal/transport/http/response"
    "github.com/hutamy/go-invoice-backend/internal/adapter/security"
    "github.com/labstack/echo/v4"
)

func JWTMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return httpresp.Response(c, http.StatusUnauthorized, "invalid token", nil)
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := security.JWTTokenService{}.Parse(tokenStr)
		if err != nil {
			return httpresp.Response(c, http.StatusUnauthorized, "invalid token", nil)
		}

		c.Set("user_id", uint(claims["user_id"].(float64)))
		return next(c)
	}
}
