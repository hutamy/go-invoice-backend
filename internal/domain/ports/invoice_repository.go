package ports

import "github.com/hutamy/go-invoice-backend/internal/domain/entity"

type InvoiceRepository interface {
	Create(invoice *entity.Invoice) error
	GetByID(id, userID uint) (*entity.Invoice, error)
	ListByUser(userID uint, page int, pageSize int, status string) ([]entity.Invoice, int64, error)
	Update(update entity.Invoice) error
	Delete(id, userID uint) error
	SoftDeleteByUserID(userID uint) error
	RestoreByUserID(userID uint) error
	UpdateStatus(id uint, userID uint, status entity.InvoiceStatus) error
	Summary(userID uint, status string) (float64, error)
}
