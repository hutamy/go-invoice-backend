package handlers

import (
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
	response "github.com/hutamy/go-invoice-backend/internal/transport/http/response"
	"github.com/hutamy/go-invoice-backend/pkg/utils"
	"github.com/labstack/echo/v4"
)

type InvoiceHandler struct {
	UseCase ports.InvoiceUseCase
}

func NewInvoiceHandler(uc ports.InvoiceUseCase) *InvoiceHandler {
	return &InvoiceHandler{
		UseCase: uc,
	}
}

type invoiceItemReq struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"` // Ensure unit price is greater than 0

}

type invoiceReq struct {
	ClientID      *uint            `json:"client_id"`
	DueDate       string           `json:"due_date" validate:"required,datetime=2006-01-02"`
	IssueDate     string           `json:"issue_date" validate:"required,datetime=2006-01-02"`
	Items         []invoiceItemReq `json:"items" validate:"required,dive"`
	Notes         string           `json:"notes"`
	InvoiceNumber string           `json:"invoice_number" validate:"required"`
	TaxRate       float64          `json:"tax_rate"`
	DeliveryFee   float64          `json:"delivery_fee"`
	ClientName    *string          `json:"client_name"`
	ClientEmail   *string          `json:"client_email"`
	ClientAddress *string          `json:"client_address"`
	ClientPhone   *string          `json:"client_phone"`
}

type senderRequest struct {
	senderRecipientRequest
	BankName          string `json:"bank_name" validate:"required"`
	BankAccountName   string `json:"bank_account_name" validate:"required"`
	BankAccountNumber string `json:"bank_account_number" validate:"required"`
}

type senderRecipientRequest struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
	Email   string `json:"email" validate:"email"`
	Phone   string `json:"phone"`
}

type invoicePublicReq struct {
	InvoiceNumber string                 `json:"invoice_number" validate:"required"`
	IssueDate     string                 `json:"issue_date" validate:"required,datetime=2006-01-02"`
	DueDate       string                 `json:"due_date" validate:"required,datetime=2006-01-02"`
	Sender        senderRequest          `json:"sender" validate:"required"`
	Recipient     senderRecipientRequest `json:"recipient" validate:"required"`
	Items         []invoiceItemReq       `json:"items,omitempty"`
	TaxRate       float64                `json:"tax_rate,omitempty"`
	Notes         string                 `json:"notes"`
	DeliveryFee   float64                `json:"delivery_fee,omitempty"`
}

func (r *invoiceReq) validate() error {
	if r.ClientID == nil {
		if r.ClientName == nil {
			return errors.New("client name is required")
		}
		if r.ClientEmail == nil {
			return errors.New("client email is required")
		}
		if r.ClientAddress == nil {
			return errors.New("client address is required")
		}
		if r.ClientPhone == nil {
			return errors.New("client phone is required")
		}
	}

	return nil
}

type statusReq struct {
	Status string `json:"status" validate:"required,oneof=DRAFT SENT PAID"`
}

// @Summary Create Invoice
// @Description  Create a new invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param request body invoiceReq true "Invoice Request"
// @Success 201 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices [post]
func (h *InvoiceHandler) CreateInvoice(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	var req invoiceReq
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	if err := req.validate(); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	dueDate, err := time.Parse(time.DateOnly, req.DueDate)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	issueDate, err := time.Parse(time.DateOnly, req.IssueDate)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	inv := &entity.Invoice{
		UserID:        userID,
		ClientID:      req.ClientID,
		InvoiceNumber: req.InvoiceNumber,
		DueDate:       dueDate,
		IssueDate:     issueDate,
		Notes:         req.Notes,
		Status:        string(entity.InvoiceStatusDraft),
		TaxRate:       req.TaxRate,
		ClientName:    req.ClientName,
		ClientEmail:   req.ClientEmail,
		ClientAddress: req.ClientAddress,
		ClientPhone:   req.ClientPhone,
		DeliveryFee:   req.DeliveryFee,
	}
	for _, it := range req.Items {
		inv.Items = append(inv.Items, entity.InvoiceItem{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
		})
	}

	if err := h.UseCase.Create(inv); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusCreated, "created", inv)
}

// @Summary Get Invoice By ID
// @Description  Get invoice by id
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices/{id} [get]
func (h *InvoiceHandler) GetInvoiceByID(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || invoiceID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	inv, err := h.UseCase.GetByID(uint(invoiceID), userID)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	if inv == nil {
		return response.Response(c, http.StatusNotFound, "not found", nil)
	}

	return response.Response(c, http.StatusOK, "ok", inv)
}

// @Summary List Invoices By User ID
// @Description  List invoices by user id
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param page query int false "Page"
// @Param page_size query int false "Page Size"
// @Param status query string false "Status"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices [get]
func (h *InvoiceHandler) ListInvoicesByUserID(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	page := utils.ParseIntDefault(c.QueryParam("page"), 1)
	size := utils.ParseIntDefault(c.QueryParam("page_size"), 10)
	status := c.QueryParam("status")
	items, total, err := h.UseCase.ListByUser(userID, page, size, status)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", map[string]any{
		"data": items,
		"pagination": map[string]any{
			"total_items": total,
			"page":        page,
			"page_size":   size,
			"total_pages": int(math.Ceil(float64(total) / float64(size))),
		},
	})
}

// @Summary Update Invoice
// @Description  Update invoice by id
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body invoiceReq true "Invoice Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices/{id} [put]
func (h *InvoiceHandler) UpdateInvoice(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || invoiceID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	var req invoiceReq
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	if err := req.validate(); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	dueDate, err := time.Parse(time.DateOnly, req.DueDate)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	issueDate, err := time.Parse(time.DateOnly, req.IssueDate)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	upd := entity.Invoice{
		ID:            uint(invoiceID),
		UserID:        userID,
		ClientID:      req.ClientID,
		InvoiceNumber: req.InvoiceNumber,
		DueDate:       dueDate,
		IssueDate:     issueDate,
		Notes:         req.Notes,
		Status:        string(entity.InvoiceStatusDraft),
		TaxRate:       req.TaxRate,
		ClientName:    req.ClientName,
		ClientEmail:   req.ClientEmail,
		ClientAddress: req.ClientAddress,
		ClientPhone:   req.ClientPhone,
	}
	for _, it := range req.Items {
		upd.Items = append(upd.Items, entity.InvoiceItem{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
		})
	}

	if err := h.UseCase.Update(upd); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "updated", nil)
}

// @Summary Delete Invoice
// @Description  Delete invoice by id
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices/{id} [delete]
func (h *InvoiceHandler) DeleteInvoice(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || invoiceID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	if err := h.UseCase.Delete(uint(invoiceID), userID); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "deleted", nil)
}

// @Summary Update Invoice Status
// @Description  Update invoice status by id
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Invoice ID"
// @Param request body statusReq true "Invoice Status Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices/{id}/status [patch]
func (h *InvoiceHandler) UpdateInvoiceStatus(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || invoiceID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	var req statusReq
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	if err := h.UseCase.UpdateStatus(uint(invoiceID), userID, entity.InvoiceStatus(req.Status)); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "updated", nil)
}

// @Summary Invoice Summary
// @Description  Invoice summary
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices/summary [get]
func (h *InvoiceHandler) Summary(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	paid, total, err := h.UseCase.Summary(userID)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", map[string]any{
		"paid":          paid,
		"total_revenue": total,
	})
}

// @Summary Download Invoice PDF
// @Description  Download invoice pdf
// @Tags Invoice
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Invoice ID"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/invoices/{id}/pdf [post]
func (h *InvoiceHandler) DownloadInvoicePDF(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	invoiceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || invoiceID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	pdf, err := h.UseCase.GeneratePDF(uint(invoiceID), userID)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return c.Blob(http.StatusOK, "application/pdf", pdf)
}

// @Summary Generate Public Invoice
// @Description  Generate public invoice
// @Tags Invoice
// @Accept json
// @Produce json
// @Param request body invoicePublicReq true "Invoice Public Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/public/invoices/generate-pdf [post]
func (h *InvoiceHandler) GeneratePublicInvoice(c echo.Context) error {
	var req invoicePublicReq
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	dueDate, err := time.Parse(time.DateOnly, req.DueDate)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	issueDate, err := time.Parse(time.DateOnly, req.IssueDate)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	inv := entity.Invoice{
		User: entity.User{
			Name:              req.Sender.Name,
			Email:             req.Sender.Email,
			Address:           req.Sender.Address,
			Phone:             req.Sender.Phone,
			BankName:          req.Sender.BankName,
			BankAccountNumber: req.Sender.BankAccountNumber,
			BankAccountName:   req.Sender.BankAccountName,
		},
		Client: entity.Client{
			Name:    req.Recipient.Name,
			Email:   req.Recipient.Email,
			Address: req.Recipient.Address,
			Phone:   req.Recipient.Phone,
		},
		InvoiceNumber: req.InvoiceNumber,
		IssueDate:     issueDate,
		DueDate:       dueDate,
		Notes:         req.Notes,
		TaxRate:       req.TaxRate,
		DeliveryFee:   req.DeliveryFee,
	}

	for _, it := range req.Items {
		inv.Items = append(inv.Items, entity.InvoiceItem{
			Description: it.Description,
			Quantity:    it.Quantity,
			UnitPrice:   it.UnitPrice,
		})
	}

	pdf, err := h.UseCase.GeneratePDFPublic(&inv)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return c.Blob(http.StatusOK, "application/pdf", pdf)
}
