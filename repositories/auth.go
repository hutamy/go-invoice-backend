package repositories

import (
	"errors"

	"github.com/hutamy/go-invoice-backend/dto"
	"github.com/hutamy/go-invoice-backend/models"
	"gorm.io/gorm"
)

type AuthRepository interface {
	CreateUser(user *models.User) error
	GetUserByEmail(email string) (*models.User, error)
	GetUserByEmailIncludingDeleted(email string) (*models.User, error)
	GetUserByID(id uint) (*models.User, error)
	UpdatePassword(id uint, password string) error
	UpdateUserProfile(req dto.UpdateUserProfileRequest) error
	UpdateUserBanking(req dto.UpdateUserBankingRequest) error
	DeleteUser(id uint) error
	RestoreUser(id uint) error
	RestoreAndUpdateUser(id uint, userData *models.User) error
	GetDeletedUserByID(id uint) (*models.User, error)
}

type authRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{db: db}
}

func (r *authRepository) CreateUser(user *models.User) error {
	return r.db.Create(&user).Error
}

func (r *authRepository) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &user, err
}

func (r *authRepository) GetUserByEmailIncludingDeleted(email string) (*models.User, error) {
	var user models.User
	err := r.db.Unscoped().Where("email = ?", email).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &user, err
}

func (r *authRepository) GetUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *authRepository) UpdatePassword(id uint, password string) error {
	user := models.User{
		Password: password,
	}

	return r.db.Model(&user).Where("id = ?", id).Update("password", password).Error
}

func (r *authRepository) UpdateUserProfile(req dto.UpdateUserProfileRequest) error {
	user := models.User{}
	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Email != nil {
		user.Email = *req.Email
	}

	if req.Address != nil {
		user.Address = *req.Address
	}

	if req.Phone != nil {
		user.Phone = *req.Phone
	}

	res := r.db.Model(&models.User{}).Where("id = ?", req.UserID).Updates(user)
	return res.Error
}

func (r *authRepository) UpdateUserBanking(req dto.UpdateUserBankingRequest) error {
	user := models.User{}
	if req.BankName != nil {
		user.BankName = *req.BankName
	}

	if req.BankAccountName != nil {
		user.BankAccountName = *req.BankAccountName
	}

	if req.BankAccountNumber != nil {
		user.BankAccountNumber = *req.BankAccountNumber
	}

	res := r.db.Model(&models.User{}).Where("id = ?", req.UserID).Updates(user)
	return res.Error
}

func (r *authRepository) DeleteUser(id uint) error {
	result := r.db.Delete(&models.User{}, id)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *authRepository) RestoreUser(id uint) error {
	result := r.db.Unscoped().Model(&models.User{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *authRepository) RestoreAndUpdateUser(id uint, userData *models.User) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	result := tx.Unscoped().Model(&models.User{}).Where("id = ?", id).Update("deleted_at", nil)
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}

	if result.RowsAffected == 0 {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	updateResult := tx.Model(&models.User{}).Where("id = ?", id).Updates(userData)
	if updateResult.Error != nil {
		tx.Rollback()
		return updateResult.Error
	}

	return tx.Commit().Error
}

func (r *authRepository) GetDeletedUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.Unscoped().Where("id = ? AND deleted_at IS NOT NULL", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
