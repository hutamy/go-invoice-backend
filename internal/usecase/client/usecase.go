package client

import (
	"errors"

	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
)

type UseCase struct {
	Repo ports.ClientRepository
}

func NewUseCase(repo ports.ClientRepository) ports.ClientUseCase {
	return &UseCase{
		Repo: repo,
	}
}

func (u *UseCase) Create(c *entity.Client) error {
	return u.Repo.Create(c)
}

func (u *UseCase) GetByID(id, userID uint) (*entity.Client, error) {
	return u.Repo.GetByID(id, userID)
}

func (u *UseCase) ListByUser(userID uint, page, pageSize int, search string) ([]entity.Client, int64, error) {
	if userID == 0 {
		return nil, 0, errors.New("unauthorized")
	}

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	return u.Repo.ListByUser(userID, page, pageSize, search)
}

func (u *UseCase) Update(update entity.Client) error {
	return u.Repo.Update(update)
}

func (u *UseCase) Delete(id uint, userID uint) error {
	return u.Repo.Delete(id, userID)
}
