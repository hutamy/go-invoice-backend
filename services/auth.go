package services

import (
	"github.com/hutamy/go-invoice-backend/dto"
	"github.com/hutamy/go-invoice-backend/models"
	"github.com/hutamy/go-invoice-backend/repositories"
	"github.com/hutamy/go-invoice-backend/utils"
	"github.com/hutamy/go-invoice-backend/utils/errors"
)

type AuthService interface {
	SignUp(req dto.SignUpRequest) (models.User, error)
	SignIn(email, password string) (models.User, error)
	GetUserByID(id uint) (*models.User, error)
	ChangePassword(req dto.ChangePasswordRequest) error
	UpdateUserBanking(req dto.UpdateUserBankingRequest) error
	UpdateUserProfile(req dto.UpdateUserProfileRequest) error
	DeactivateUser(id uint) error
}

type authService struct {
	authRepo    repositories.AuthRepository
	clientRepo  repositories.ClientRepository
	invoiceRepo repositories.InvoiceRepository
}

func NewAuthService(
	authRepo repositories.AuthRepository,
	clientRepo repositories.ClientRepository,
	invoiceRepo repositories.InvoiceRepository,
) AuthService {
	return &authService{
		authRepo:    authRepo,
		clientRepo:  clientRepo,
		invoiceRepo: invoiceRepo,
	}
}

func (s *authService) SignUp(req dto.SignUpRequest) (models.User, error) {
	existingUser, err := s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		return models.User{}, err
	}

	if existingUser != nil {
		return models.User{}, errors.ErrUserAlreadyExists
	}

	deletedUser, err := s.authRepo.GetUserByEmailIncludingDeleted(req.Email)
	if err != nil {
		return models.User{}, err
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return models.User{}, err
	}

	// If a soft-deleted user exists, restore and update their data
	if deletedUser != nil && deletedUser.DeletedAt.Valid {
		updatedUserData := &models.User{
			Name:              req.Name,
			Password:          string(hashedPassword),
			Address:           req.Address,
			Phone:             req.Phone,
			BankName:          req.BankName,
			BankAccountName:   req.BankAccountName,
			BankAccountNumber: req.BankAccountNumber,
		}

		if err := s.authRepo.RestoreAndUpdateUser(deletedUser.ID, updatedUserData); err != nil {
			return models.User{}, err
		}

		if err := s.clientRepo.RestoreAllByUserID(deletedUser.ID); err != nil {
			return models.User{}, err
		}

		if err := s.invoiceRepo.RestoreAllByUserID(deletedUser.ID); err != nil {
			return models.User{}, err
		}

		restoredUser, err := s.authRepo.GetUserByEmail(req.Email)
		if err != nil {
			return models.User{}, err
		}

		if restoredUser == nil {
			return models.User{}, errors.ErrUserNotFound
		}

		return *restoredUser, nil
	}

	// Create a new user if no soft-deleted user exists
	user := &models.User{
		Name:              req.Name,
		Email:             req.Email,
		Password:          string(hashedPassword),
		Address:           req.Address,
		Phone:             req.Phone,
		BankName:          req.BankName,
		BankAccountName:   req.BankAccountName,
		BankAccountNumber: req.BankAccountNumber,
	}

	if err := s.authRepo.CreateUser(user); err != nil {
		return models.User{}, err
	}

	user, err = s.authRepo.GetUserByEmail(req.Email)
	if err != nil {
		return models.User{}, err
	}

	return *user, nil
}

func (s *authService) SignIn(email, password string) (models.User, error) {
	user, err := s.authRepo.GetUserByEmail(email)
	if err != nil {
		return models.User{}, err
	}

	// If no active user found, check if user exists but is soft-deleted
	if user == nil {
		deletedUser, err := s.authRepo.GetUserByEmailIncludingDeleted(email)
		if err != nil {
			return models.User{}, err
		}

		if deletedUser != nil && deletedUser.DeletedAt.Valid {
			return models.User{}, errors.ErrAccountDeactivated
		}

		// No user found at all (active or deleted)
		return models.User{}, errors.ErrLoginFailed
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return models.User{}, errors.ErrLoginFailed
	}

	return *user, nil
}

func (s *authService) GetUserByID(id uint) (*models.User, error) {
	return s.authRepo.GetUserByID(id)
}

func (s *authService) ChangePassword(req dto.ChangePasswordRequest) error {
	user, err := s.authRepo.GetUserByID(req.UserID)
	if err != nil {
		return err
	}

	if user == nil {
		return errors.ErrUserNotFound
	}

	if !utils.CheckPasswordHash(req.OldPassword, user.Password) {
		return errors.ErrInvalidOldPassword
	}

	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}

	return s.authRepo.UpdatePassword(req.UserID, string(hashedPassword))
}

func (s *authService) UpdateUserBanking(req dto.UpdateUserBankingRequest) error {
	return s.authRepo.UpdateUserBanking(req)
}

func (s *authService) UpdateUserProfile(req dto.UpdateUserProfileRequest) error {
	return s.authRepo.UpdateUserProfile(req)
}

func (s *authService) DeactivateUser(id uint) error {
	user, err := s.authRepo.GetUserByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.ErrUserNotFound
	}

	if err := s.clientRepo.SoftDeleteAllByUserID(id); err != nil {
		return err
	}

	if err := s.invoiceRepo.SoftDeleteAllByUserID(id); err != nil {
		return err
	}

	return s.authRepo.DeleteUser(id)
}
