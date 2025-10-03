package entity

type InvoiceItem struct {
	ID          uint    `json:"id"`
	InvoiceID   uint    `json:"invoice_id"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Total       float64 `json:"total"`
}
