package routes

import (
	"github.com/hutamy/go-invoice-backend/controllers"
	_ "github.com/hutamy/go-invoice-backend/docs"
	"github.com/hutamy/go-invoice-backend/middleware"
	"github.com/hutamy/go-invoice-backend/repositories"
	"github.com/hutamy/go-invoice-backend/services"
	"github.com/hutamy/go-invoice-backend/utils"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
	"gorm.io/gorm"
)

func InitRoutes(e *echo.Echo, db *gorm.DB) {
	authRepo := repositories.NewAuthRepository(db)
	clientRepo := repositories.NewClientRepository(db)
	invoiceRepo := repositories.NewInvoiceRepository(db)

	authService := services.NewAuthService(authRepo, clientRepo, invoiceRepo)
	authController := controllers.NewAuthController(authService)

	clientService := services.NewClientService(clientRepo)
	clientController := controllers.NewClientController(clientService)

	invoiceService := services.NewInvoiceService(invoiceRepo, clientRepo, authRepo)
	invoiceController := controllers.NewInvoiceController(invoiceService)

	e.GET("/", func(c echo.Context) error {
		return utils.Response(c, 200, "Welcome to Go Invoice API", nil)
	})
	e.GET("/health", func(c echo.Context) error {
		return utils.Response(c, 200, "Go Invoice API is running", nil)
	})
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := e.Group("/v1")
	public := v1.Group("/public")

	authRoutes := public.Group("/auth")
	authRoutes.POST("/sign-up", authController.SignUp)
	authRoutes.POST("/sign-in", authController.SignIn)

	publicInvoiceRoutes := public.Group("/invoices")
	publicInvoiceRoutes.POST("/generate-pdf", invoiceController.GeneratePublicInvoice)

	protected := v1.Group("/protected")
	protected.Use(middleware.JWTMiddleware)

	protected.GET("/me", authController.Me)
	protected.PUT("/me/banking", authController.UpdateUserBanking)
	protected.PUT("/me/profile", authController.UpdateUserProfile)
	protected.POST("/me/change-password", authController.ChangePassword)
	protected.POST("/me/deactivate", authController.DeactivateUser)

	authPrivateRoutes := protected.Group("/auth")
	authPrivateRoutes.POST("/refresh-token", authController.RefreshToken)

	clientRoutes := protected.Group("/clients")
	clientRoutes.POST("", clientController.CreateClient)
	clientRoutes.GET("", clientController.GetAllClients)
	clientRoutes.GET("/:id", clientController.GetClientByID)
	clientRoutes.PUT("/:id", clientController.UpdateClient)
	clientRoutes.DELETE("/:id", clientController.DeleteClient)

	protectedInvoiceRoutes := protected.Group("/invoices")
	protectedInvoiceRoutes.GET("/summary", invoiceController.InvoiceSummary)
	protectedInvoiceRoutes.POST("", invoiceController.CreateInvoice)
	protectedInvoiceRoutes.GET("/:id", invoiceController.GetInvoiceByID)
	protectedInvoiceRoutes.PUT("/:id", invoiceController.UpdateInvoice)
	protectedInvoiceRoutes.DELETE("/:id", invoiceController.DeleteInvoice)
	protectedInvoiceRoutes.GET("", invoiceController.ListInvoicesByUserID)
	protectedInvoiceRoutes.PATCH("/:id/status", invoiceController.UpdateInvoiceStatus)
	protectedInvoiceRoutes.POST("/:id/pdf", invoiceController.DownloadInvoicePDF)
}
