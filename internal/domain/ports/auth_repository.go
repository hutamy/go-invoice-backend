package ports

import "github.com/hutamy/go-invoice-backend/internal/domain/entity"

type AuthRepository interface {
	CreateUser(user *entity.User) error
	GetUserByEmail(email string) (*entity.User, error)
	GetUserByID(id uint) (*entity.User, error)
	UpdatePassword(id uint, password string) error
	UpdateUserProfile(userID uint, update entity.User) error
	UpdateUserBanking(userID uint, update entity.User) error
	DeleteUser(id uint) error
	RestoreUser(user *entity.User) error
}
