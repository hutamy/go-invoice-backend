package ports

import "github.com/hutamy/go-invoice-backend/internal/domain/entity"

type InvoiceUseCase interface {
	Create(invoice *entity.Invoice) error
	GetByID(id, userID uint) (*entity.Invoice, error)
	ListByUser(userID uint, page int, pageSize int, status string) ([]entity.Invoice, int64, error)
	Update(update entity.Invoice) error
	Delete(id uint, userID uint) error
	UpdateStatus(id uint, userID uint, status entity.InvoiceStatus) error
	Summary(userID uint) (paid, revenue float64, err error)
	GeneratePDFPublic(invoice *entity.Invoice) ([]byte, error)
	GeneratePDF(id, userID uint) ([]byte, error)
}
