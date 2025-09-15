package dto

type CreateClientRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Address string `json:"address" validate:"required"`
	Phone   string `json:"phone" validate:"required"`
	UserID  uint   `json:"-"`
}

type UpdateClientRequest struct {
	Name    *string `json:"name" validate:"omitempty"`
	Email   *string `json:"email" validate:"omitempty,email"`
	Address *string `json:"address" validate:"omitempty"`
	Phone   *string `json:"phone" validate:"omitempty"`
	ID      uint    `param:"id" validate:"required"`
	UserID  uint    `json:"-"`
}

type PaginationRequest struct {
	Page     int    `query:"page" validate:"min=1"`
	PageSize int    `query:"page_size" validate:"min=1,max=100"`
	Search   string `query:"search"`
}

type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

type GetClientsRequest struct {
	UserID uint `json:"-"`
	PaginationRequest
}
