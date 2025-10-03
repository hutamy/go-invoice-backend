package ports

import "github.com/hutamy/go-invoice-backend/internal/domain/entity"

type ClientRepository interface {
	Create(client *entity.Client) error
	GetByID(id, userID uint) (*entity.Client, error)
	ListByUser(userID uint, page int, pageSize int, search string) ([]entity.Client, int64, error)
	Update(update entity.Client) error
	Delete(id, userID uint) error
	SoftDeleteByUserID(userID uint) error
	RestoreByUserID(userID uint) error
}
