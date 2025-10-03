package entity

import (
	"time"

	"gorm.io/gorm"
)

type InvoiceStatus string

const (
	InvoiceStatusDraft InvoiceStatus = "DRAFT"
	InvoiceStatusSent  InvoiceStatus = "SENT"
	InvoiceStatusPaid  InvoiceStatus = "PAID"
)

type Invoice struct {
	ID            uint           `json:"id"`
	UserID        uint           `json:"user_id"`
	ClientID      *uint          `json:"client_id"`
	ClientName    *string        `json:"client_name"`
	ClientEmail   *string        `json:"client_email"`
	ClientAddress *string        `json:"client_address"`
	ClientPhone   *string        `json:"client_phone"`
	InvoiceNumber string         `json:"invoice_number"`
	IssueDate     time.Time      `json:"issue_date"`
	DueDate       time.Time      `json:"due_date"`
	Status        string         `json:"status"`
	Notes         string         `json:"notes"`
	Subtotal      float64        `json:"subtotal"`
	Tax           float64        `json:"tax"`
	TaxRate       float64        `json:"tax_rate"`
	DeliveryFee   float64        `json:"delivery_fee"`
	Total         float64        `json:"total"`
	Items         []InvoiceItem  `json:"items"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-"`

	// Relationship
	User   User
	Client Client
}
