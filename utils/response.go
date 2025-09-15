package utils

import "github.com/labstack/echo/v4"

type GenericResponse struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func Response(c echo.Context, status int, message string, data interface{}) error {
	return c.JSON(status, GenericResponse{
		Status:  status,
		Message: message,
		Data:    data,
	})
}
