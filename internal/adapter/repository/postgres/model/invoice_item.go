package model

type InvoiceItem struct {
	ID          uint    `json:"id" gorm:"primaryKey"`
	InvoiceID   uint    `json:"invoice_id" gorm:"not null;index"`
	Description string  `json:"description" gorm:"type:text"`
	Quantity    int     `json:"quantity" gorm:"not null;default:1"`
	UnitPrice   float64 `json:"unit_price" gorm:"not null;default:0"`
	Total       float64 `json:"total" gorm:"not null;default:0"`
}
