package utils

import (
	"math"

	"github.com/hutamy/go-invoice-backend/dto"
)

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Data       interface{}            `json:"data"`
	Pagination dto.PaginationResponse `json:"pagination"`
}

// CalculatePagination calculates pagination metadata
func CalculatePagination(page, pageSize int, totalItems int64) dto.PaginationResponse {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))

	return dto.PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
	}
}

// PaginatedData creates a paginated response with data and pagination info
func PaginatedData(data interface{}, pagination dto.PaginationResponse) PaginatedResponse {
	return PaginatedResponse{
		Data:       data,
		Pagination: pagination,
	}
}
