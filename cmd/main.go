// @title Go Invoice API
// @version 1.0
// @description API documentation for your invoice app.
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

package main

import (
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/hutamy/go-invoice-backend/config"
	"github.com/hutamy/go-invoice-backend/routes"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func main() {
	log.Println("Starting application...")

	cfg := config.LoadEnv()
	log.Printf("Loaded config - Port: %d, SkipMigrate: %t", cfg.Port, cfg.SkipMigrate)

	log.Println("Initializing database connection...")
	db := config.InitDB(cfg.DatabaseURL)
	log.Println("Database connection successful!")

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	log.Println("Initializing routes...")
	routes.InitRoutes(e, db)
	log.Println("Routes initialized successfully!")

	log.Printf("Starting server on port: %d", cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)))
}
