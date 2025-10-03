package ports

import "github.com/hutamy/go-invoice-backend/internal/domain/entity"

type AuthUseCase interface {
	SignUp(user *entity.User) (access, refresh string, err error)
	SignIn(email, password string) (accessToken, refreshToken string, err error)
	Me(userID uint) (*entity.User, error)
	UpdateUserProfile(userID uint, update entity.User) error
	UpdateUserBanking(userID uint, update entity.User) error
	ChangePassword(userID uint, oldPassword, newPassword string) error
	DeactivateUser(userID uint) error
	RefreshToken(token string) (access, refresh string, err error)
}
