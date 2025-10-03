package http

import (
	"github.com/hutamy/go-invoice-backend/internal/transport/http/handlers"
	"github.com/hutamy/go-invoice-backend/middleware"
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type RouterDeps struct {
	Auth    *handlers.AuthHandler
	Client  *handlers.ClientHandler
	Invoice *handlers.InvoiceHandler
}

func RegisterRoutes(e *echo.Echo, deps RouterDeps) {
	e.GET("/", deps.Auth.Health)
	e.GET("/health", deps.Auth.Health)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	v1 := e.Group("/v1")
	public := v1.Group("/public")
	authPublic := public.Group("/auth")
	authPublic.POST("/sign-up", deps.Auth.SignUp)
	authPublic.POST("/sign-in", deps.Auth.SignIn)

	protected := v1.Group("/protected")
	protected.Use(middleware.JWTMiddleware)
	protected.GET("/me", deps.Auth.Me)
	protected.PUT("/me/banking", deps.Auth.UpdateBanking)
	protected.PUT("/me/profile", deps.Auth.UpdateProfile)
	protected.POST("/me/change-password", deps.Auth.ChangePassword)
	protected.POST("/me/deactivate", deps.Auth.DeactivateUser)
	protected.POST("/auth/refresh-token", deps.Auth.RefreshToken)

	clientRoutes := protected.Group("/clients")
	clientRoutes.POST("", deps.Client.CreateClient)
	clientRoutes.GET("", deps.Client.GetAllClients)
	clientRoutes.GET("/:id", deps.Client.GetClientByID)
	clientRoutes.PUT("/:id", deps.Client.UpdateClient)
	clientRoutes.DELETE("/:id", deps.Client.DeleteClient)

	invoiceRoutes := protected.Group("/invoices")
	invoiceRoutes.GET("/summary", deps.Invoice.Summary)
	invoiceRoutes.POST("", deps.Invoice.CreateInvoice)
	invoiceRoutes.GET("/:id", deps.Invoice.GetInvoiceByID)
	invoiceRoutes.PUT("/:id", deps.Invoice.UpdateInvoice)
	invoiceRoutes.DELETE("/:id", deps.Invoice.DeleteInvoice)
	invoiceRoutes.GET("", deps.Invoice.ListInvoicesByUserID)
	invoiceRoutes.PATCH("/:id/status", deps.Invoice.UpdateInvoiceStatus)
	invoiceRoutes.POST("/:id/pdf", deps.Invoice.DownloadInvoicePDF)

	publicInvoices := public.Group("/invoices")
	publicInvoices.POST("/generate-pdf", deps.Invoice.GeneratePublicInvoice)
}
