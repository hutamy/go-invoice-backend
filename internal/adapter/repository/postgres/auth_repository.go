package postgres

import (
	"errors"

	"github.com/hutamy/go-invoice-backend/internal/adapter/mapper"
	pmodel "github.com/hutamy/go-invoice-backend/internal/adapter/repository/postgres/model"
	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) ports.AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUser(user *entity.User) error {
	m := mapper.UserToModel(user)
	return r.db.Create(m).Error
}

func (r *AuthRepository) GetUserByEmail(email string) (*entity.User, error) {
	var m pmodel.User
	err := r.db.Unscoped().Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return mapper.UserFromModel(&m), nil
}

func (r *AuthRepository) GetUserByID(id uint) (*entity.User, error) {
	var m pmodel.User
	err := r.db.First(&m, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return mapper.UserFromModel(&m), nil
}

func (r *AuthRepository) UpdatePassword(id uint, password string) error {
	return r.db.Model(&pmodel.User{}).
		Where("id = ?", id).
		Update("password", password).Error
}

func (r *AuthRepository) UpdateUserProfile(userID uint, update entity.User) error {
	updates := map[string]any{
		"name":    update.Name,
		"email":   update.Email,
		"address": update.Address,
		"phone":   update.Phone,
	}
	return r.db.Model(&pmodel.User{}).
		Where("id = ?", userID).
		Updates(updates).Error
}

func (r *AuthRepository) UpdateUserBanking(userID uint, update entity.User) error {
	updates := map[string]any{
		"bank_name":           update.BankName,
		"bank_account_name":   update.BankAccountName,
		"bank_account_number": update.BankAccountNumber,
	}
	return r.db.Model(&pmodel.User{}).
		Where("id = ?", userID).
		Updates(updates).Error
}

func (r *AuthRepository) DeleteUser(id uint) error {
	res := r.db.Delete(&pmodel.User{}, id)
	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *AuthRepository) RestoreUser(user *entity.User) error {
	m := mapper.UserToModel(user)
	updates := map[string]any{
		"deleted_at":          nil,
		"name":                m.Name,
		"email":               m.Email,
		"password":            m.Password,
		"address":             m.Address,
		"phone":               m.Phone,
		"bank_name":           m.BankName,
		"bank_account_name":   m.BankAccountName,
		"bank_account_number": m.BankAccountNumber,
	}

	res := r.db.Unscoped().Model(&pmodel.User{}).
		Where("id = ? AND deleted_at IS NOT NULL", m.ID).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
