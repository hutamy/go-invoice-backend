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
	_ "github.com/hutamy/go-invoice-backend/docs"
	pgrepo "github.com/hutamy/go-invoice-backend/internal/adapter/repository/postgres"
	"github.com/hutamy/go-invoice-backend/internal/adapter/security"
	ht "github.com/hutamy/go-invoice-backend/internal/transport/http"
	"github.com/hutamy/go-invoice-backend/internal/transport/http/handlers"
	authuc "github.com/hutamy/go-invoice-backend/internal/usecase/auth"
	clientuc "github.com/hutamy/go-invoice-backend/internal/usecase/client"
	invoiceuc "github.com/hutamy/go-invoice-backend/internal/usecase/invoice"
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
	db := config.InitDB(cfg.DatabaseURL, cfg.Schema)
	log.Println("Database connection successful!")

	e := echo.New()

	// Register custom validator
	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	log.Println("Initializing Clean Architecture router...")
	// Wire repositories (direct adapters)
	authRepo := pgrepo.NewAuthRepository(db)
	clientRepo := pgrepo.NewClientRepository(db)
	invoiceRepo := pgrepo.NewInvoiceRepository(db)

	// Security adapters
	hasher := security.NewBcryptHasher()
	tokens := security.NewJWTTokenService()

	// Wire use cases
	authUC := authuc.NewUseCase(authRepo, clientRepo, invoiceRepo, hasher, tokens)
	clientUC := clientuc.NewUseCase(clientRepo)
	invoiceUC := invoiceuc.NewUseCase(invoiceRepo, clientRepo, authRepo)

	// Handlers
	authHandler := handlers.NewAuthHandler(authUC)
	clientHandler := handlers.NewClientHandler(clientUC)
	invoiceHandler := handlers.NewInvoiceHandler(invoiceUC)

	// Register routes
	ht.RegisterRoutes(e, ht.RouterDeps{
		Auth:    authHandler,
		Client:  clientHandler,
		Invoice: invoiceHandler,
	})

	log.Printf("Starting server on port: %d", cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("0.0.0.0:%d", cfg.Port)))
}
