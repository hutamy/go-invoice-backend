package services

import (
	"bytes"
	"context"
	"html/template"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/dustin/go-humanize"
	"github.com/hutamy/go-invoice-backend/dto"
	"github.com/hutamy/go-invoice-backend/models"
	"github.com/hutamy/go-invoice-backend/repositories"
	"github.com/hutamy/go-invoice-backend/utils"
)

type InvoiceService interface {
	CreateInvoice(invoice *models.Invoice) error
	GetInvoiceByID(id uint) (*models.Invoice, error)
	ListInvoiceByUserID(userID uint) ([]models.Invoice, error)
	ListInvoiceByUserIDWithPagination(req dto.GetInvoicesRequest) (utils.PaginatedResponse, error)
	UpdateInvoice(id uint, req *dto.UpdateInvoiceRequest) error
	GenerateInvoicePDF(invoiceID uint) ([]byte, error)
	GeneratePublicInvoicePDF(req dto.GeneratePublicInvoiceRequest) ([]byte, error)
	DeleteInvoice(id uint) error
	UpdateInvoiceStatus(id uint, status string) error
	InvoiceSummary(userID uint) (dto.SummaryInvoice, error)
}

type invoiceService struct {
	invoiceRepo repositories.InvoiceRepository
	clientRepo  repositories.ClientRepository
	authRepo    repositories.AuthRepository
}

func NewInvoiceService(
	invoiceRepo repositories.InvoiceRepository,
	clientRepo repositories.ClientRepository,
	authRepo repositories.AuthRepository,
) InvoiceService {
	return &invoiceService{
		invoiceRepo: invoiceRepo,
		clientRepo:  clientRepo,
		authRepo:    authRepo,
	}
}

func (s *invoiceService) CreateInvoice(invoice *models.Invoice) error {
	var subtotal float64
	for i, item := range invoice.Items {
		total := float64(item.Quantity) * item.UnitPrice
		invoice.Items[i].Total = total
		subtotal += total
	}

	invoice.Subtotal = subtotal
	invoice.Tax = invoice.TaxRate * subtotal / 100
	invoice.Total = invoice.Subtotal + invoice.Tax
	invoice.Status = "draft" // Default status for new invoices
	return s.invoiceRepo.CreateInvoice(invoice)
}

func (s *invoiceService) GetInvoiceByID(id uint) (*models.Invoice, error) {
	return s.invoiceRepo.GetInvoiceByID(id)
}

func (s *invoiceService) ListInvoiceByUserID(userID uint) ([]models.Invoice, error) {
	return s.invoiceRepo.ListInvoiceByUserID(userID)
}

func (s *invoiceService) ListInvoiceByUserIDWithPagination(req dto.GetInvoicesRequest) (utils.PaginatedResponse, error) {
	// Set default values for pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	invoices, totalItems, err := s.invoiceRepo.ListInvoiceByUserIDWithPagination(req)
	if err != nil {
		return utils.PaginatedResponse{}, err
	}

	pagination := utils.CalculatePagination(req.Page, req.PageSize, totalItems)
	return utils.PaginatedData(invoices, pagination), nil
}

func (s *invoiceService) UpdateInvoice(id uint, req *dto.UpdateInvoiceRequest) error {
	return s.invoiceRepo.UpdateInvoice(id, req)
}

func (s *invoiceService) GenerateInvoicePDF(invoiceID uint) ([]byte, error) {
	var err error
	var invoice *models.Invoice
	invoice, err = s.invoiceRepo.GetInvoiceByID(invoiceID)
	if err != nil {
		return nil, err
	}

	var client *models.Client
	if invoice.ClientID != nil {
		client, err = s.clientRepo.GetClientByID(*invoice.ClientID, invoice.UserID)
		if err != nil {
			return nil, err
		}
	} else {
		client = &models.Client{
			ID:      0,
			Name:    *invoice.ClientName,
			Email:   *invoice.ClientEmail,
			Address: *invoice.ClientAddress,
			Phone:   *invoice.ClientPhone,
		}
	}

	user, err := s.authRepo.GetUserByID(invoice.UserID)
	if err != nil {
		return nil, err
	}

	// Load HTML template
	htmlContent, err := s.generateHTMLContent(invoice, client, user)
	if err != nil {
		return nil, err
	}

	return s.generatePdf(htmlContent)
}

func (s *invoiceService) GeneratePublicInvoicePDF(req dto.GeneratePublicInvoiceRequest) ([]byte, error) {
	user := &models.User{
		Name:              req.Sender.Name,
		Email:             req.Sender.Email,
		Address:           req.Sender.Address,
		Phone:             req.Sender.Phone,
		BankName:          req.Sender.BankName,
		BankAccountName:   req.Sender.BankAccountName,
		BankAccountNumber: req.Sender.BankAccountNumber,
	}

	issueDate, _ := time.Parse(time.DateOnly, req.IssueDate)
	dueDate, _ := time.Parse(time.DateOnly, req.DueDate)
	invoice := &models.Invoice{
		InvoiceNumber: req.InvoiceNumber,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		Notes:         req.Notes,
		TaxRate:       req.TaxRate,
		Items:         make([]models.InvoiceItem, len(req.Items)),
	}

	for i, item := range req.Items {
		total := float64(item.Quantity) * item.UnitPrice
		invoice.Subtotal += total
		invoice.Items[i] = models.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Total:       total,
		}
	}

	invoice.Tax = invoice.TaxRate * invoice.Subtotal / 100
	invoice.Total = invoice.Subtotal + invoice.Tax
	client := &models.Client{
		Name:    req.Recipient.Name,
		Email:   req.Recipient.Email,
		Address: req.Recipient.Address,
		Phone:   req.Recipient.Phone,
	}

	// Load HTML template
	htmlContent, err := s.generateHTMLContent(invoice, client, user)
	if err != nil {
		return nil, err
	}

	return s.generatePdf(htmlContent)
}

func (s *invoiceService) generateHTMLContent(invoice *models.Invoice, client *models.Client, user *models.User) (string, error) {
	// Load HTML template
	funcMap := template.FuncMap{
		"humanize": func(value float64) string {
			return humanize.CommafWithDigits(value, 2)
		},
	}
	tmpl := template.New("invoice.html").Funcs(funcMap)
	tmpl, err := tmpl.ParseFiles("templates/invoice.html")
	if err != nil {
		return "", err
	}

	var htmlBuf bytes.Buffer
	err = tmpl.Execute(&htmlBuf, map[string]interface{}{
		"Invoice": invoice,
		"Client":  client,
		"User":    user,
	})
	if err != nil {
		return "", err
	}

	return htmlBuf.String(), nil
}

func (s *invoiceService) generatePdf(htmlContent string) ([]byte, error) {
	// Setup headless browser
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`document.documentElement.innerHTML = `+strconv.Quote(htmlContent), nil).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, err
	}

	return pdfBuf, nil
}

func (s *invoiceService) DeleteInvoice(id uint) error {
	return s.invoiceRepo.DeleteInvoice(id)
}

func (s *invoiceService) UpdateInvoiceStatus(id uint, status string) error {
	return s.invoiceRepo.UpdateInvoiceStatus(id, status)
}

func (s *invoiceService) InvoiceSummary(userID uint) (dto.SummaryInvoice, error) {
	summary, err := s.invoiceRepo.InvoiceSummary(userID)
	if err != nil {
		return dto.SummaryInvoice{}, err
	}

	return summary, nil
}
