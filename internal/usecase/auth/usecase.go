package auth

import (
	"errors"
	"time"

	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
)

type UseCase struct {
	AuthRepo    ports.AuthRepository
	ClientRepo  ports.ClientRepository
	InvoiceRepo ports.InvoiceRepository
	Hasher      ports.PasswordHasher
	Tokens      ports.TokenService
}

func NewUseCase(
	authRepo ports.AuthRepository,
	clientRepo ports.ClientRepository,
	invoiceRepo ports.InvoiceRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenService,
) ports.AuthUseCase {
	return &UseCase{
		AuthRepo:    authRepo,
		ClientRepo:  clientRepo,
		InvoiceRepo: invoiceRepo,
		Hasher:      hasher,
		Tokens:      tokens,
	}
}

func (u *UseCase) SignUp(user *entity.User) (access, refresh string, err error) {
	exist, err := u.AuthRepo.GetUserByEmail(user.Email)
	if err != nil {
		return "", "", err
	}

	hash, err := u.Hasher.Hash(user.Password)
	if err != nil {
		return "", "", err
	}

	user.Password = hash
	if exist != nil {
		if !exist.IsDeleted {
			return "", "", errors.New("email already in use")
		}

		// restore accocunt
		user.ID = exist.ID
		if err := u.AuthRepo.RestoreUser(user); err != nil {
			return "", "", err
		}

		if err := u.ClientRepo.RestoreByUserID(user.ID); err != nil {
			return "", "", err
		}

		if err := u.InvoiceRepo.RestoreByUserID(user.ID); err != nil {
			return "", "", err
		}
	} else {
		if err := u.AuthRepo.CreateUser(user); err != nil {
			return "", "", err
		}
	}

	access, err = u.Tokens.Generate(user.ID, 15*time.Minute)
	if err != nil {
		return "", "", err
	}

	refresh, err = u.Tokens.Generate(user.ID, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (u *UseCase) SignIn(email, password string) (access, refresh string, err error) {
	user, err := u.AuthRepo.GetUserByEmail(email)
	if err != nil {
		return "", "", err
	}

	if user == nil {
		return "", "", errors.New("invalid credentials")
	}

	if !u.Hasher.Compare(user.Password, password) {
		return "", "", errors.New("invalid credentials")
	}

	access, err = u.Tokens.Generate(user.ID, 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	refresh, err = u.Tokens.Generate(user.ID, 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (u *UseCase) RefreshToken(token string) (access, refresh string, err error) {
	claims, err := u.Tokens.Parse(token)
	if err != nil {
		return "", "", err
	}

	access, err = u.Tokens.Generate(uint(claims["user_id"].(float64)), 24*time.Hour)
	if err != nil {
		return "", "", err
	}

	refresh, err = u.Tokens.Generate(uint(claims["user_id"].(float64)), 7*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

func (u *UseCase) Me(userID uint) (*entity.User, error) {
	user, err := u.AuthRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}

func (u *UseCase) UpdateUserProfile(userID uint, update entity.User) error {
	return u.AuthRepo.UpdateUserProfile(userID, update)
}

func (u *UseCase) UpdateUserBanking(userID uint, update entity.User) error {
	return u.AuthRepo.UpdateUserBanking(userID, update)
}

func (u *UseCase) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := u.AuthRepo.GetUserByID(userID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.New("user not found")
	}

	if !u.Hasher.Compare(user.Password, oldPassword) {
		return errors.New("invalid credentials")
	}

	hashed, err := u.Hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	return u.AuthRepo.UpdatePassword(userID, hashed)
}

func (u *UseCase) DeactivateUser(userID uint) error {
	// soft delete clients and invoice so it can be restored
	if err := u.ClientRepo.SoftDeleteByUserID(userID); err != nil {
		return err
	}

	if err := u.InvoiceRepo.SoftDeleteByUserID(userID); err != nil {
		return err
	}

	return u.AuthRepo.DeleteUser(userID)
}
