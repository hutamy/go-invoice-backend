package dto

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

type InvoiceItemRequest struct {
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"` // Ensure unit price is greater than 0
}

type CreateInvoiceRequest struct {
	ClientID      *uint                `json:"client_id"`
	DueDate       string               `json:"due_date" validate:"required,datetime=2006-01-02"`
	IssueDate     string               `json:"issue_date" validate:"required,datetime=2006-01-02"`
	Items         []InvoiceItemRequest `json:"items" validate:"required,dive"`
	Notes         string               `json:"notes"`
	InvoiceNumber string               `json:"invoice_number" validate:"required"`
	TaxRate       float64              `json:"tax_rate"`
	ClientName    *string              `json:"client_name"`
	ClientEmail   *string              `json:"client_email"`
	ClientAddress *string              `json:"client_address"`
	ClientPhone   *string              `json:"client_phone"`
}

func (c *CreateInvoiceRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	if c.ClientID == nil {
		if c.ClientName == nil || *c.ClientName == "" {
			return errors.New("client_name is required when client_id is null")
		}
		if c.ClientEmail == nil || *c.ClientEmail == "" {
			return errors.New("client_email is required when client_id is null")
		}
		if c.ClientAddress == nil || *c.ClientAddress == "" {
			return errors.New("client_address is required when client_id is null")
		}
		if c.ClientPhone == nil || *c.ClientPhone == "" {
			return errors.New("client_phone is required when client_id is null")
		}
	}
	return nil
}

type InvoiceItemUpdateRequest struct {
	ID          *uint   `json:"id,omitempty"`
	Description string  `json:"description" validate:"required"`
	Quantity    int     `json:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" validate:"required,gt=0"` // Ensure unit price is greater than 0
}

type UpdateInvoiceRequest struct {
	ClientID      *uint                      `json:"client_id,omitempty"`
	DueDate       *string                    `json:"due_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	IssueDate     *string                    `json:"issue_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
	Notes         *string                    `json:"notes,omitempty"`
	Status        *string                    `json:"status,omitempty"`
	TaxRate       *float64                   `json:"tax_rate,omitempty"`
	InvoiceNumber *string                    `json:"invoice_number,omitempty"`
	Items         []InvoiceItemUpdateRequest `json:"items,omitempty"`
	ClientName    *string                    `json:"client_name,omitempty"`
	ClientEmail   *string                    `json:"client_email,omitempty"`
	ClientAddress *string                    `json:"client_address,omitempty"`
	ClientPhone   *string                    `json:"client_phone,omitempty"`
}

func (c *UpdateInvoiceRequest) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return err
	}
	if c.ClientID == nil {
		if c.ClientName == nil || *c.ClientName == "" {
			return errors.New("client_name is required when client_id is null")
		}
		if c.ClientEmail == nil || *c.ClientEmail == "" {
			return errors.New("client_email is required when client_id is null")
		}
		if c.ClientAddress == nil || *c.ClientAddress == "" {
			return errors.New("client_address is required when client_id is null")
		}
		if c.ClientPhone == nil || *c.ClientPhone == "" {
			return errors.New("client_phone is required when client_id is null")
		}
	}
	return nil
}

type GeneratePublicInvoiceRequest struct {
	InvoiceNumber string                     `json:"invoice_number" validate:"required"`
	IssueDate     string                     `json:"issue_date" validate:"required,datetime=2006-01-02"`
	DueDate       string                     `json:"due_date" validate:"required,datetime=2006-01-02"`
	Sender        SenderRequest              `json:"sender" validate:"required"`
	Recipient     SenderRecipientRequest     `json:"recipient" validate:"required"`
	Items         []InvoiceItemUpdateRequest `json:"items,omitempty"`
	TaxRate       float64                    `json:"tax_rate,omitempty"`
	Notes         string                     `json:"notes"`
}

type SenderRequest struct {
	SenderRecipientRequest
	BankName          string `json:"bank_name" validate:"required"`
	BankAccountName   string `json:"bank_account_name" validate:"required"`
	BankAccountNumber string `json:"bank_account_number" validate:"required"`
}

type SenderRecipientRequest struct {
	Name    string `json:"name" validate:"required"`
	Address string `json:"address" validate:"required"`
	Email   string `json:"email" validate:"email"`
	Phone   string `json:"phone"`
}

type UpdateInvoiceStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type SummaryInvoice struct {
	Paid         float64 `json:"paid"`
	TotalRevenue float64 `json:"total_revenue"`
}

type GetInvoicesRequest struct {
	UserID uint `json:"-"`
	PaginationRequest
	Status string `query:"status"` // Filter by status (draft, open, paid, past_due)
}
