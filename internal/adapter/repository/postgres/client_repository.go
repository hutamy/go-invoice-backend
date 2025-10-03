package postgres

import (
	"errors"

	"github.com/hutamy/go-invoice-backend/internal/adapter/mapper"
	"github.com/hutamy/go-invoice-backend/internal/adapter/repository/postgres/model"
	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
	"gorm.io/gorm"
)

type ClientRepository struct {
	db *gorm.DB
}

func NewClientRepository(db *gorm.DB) ports.ClientRepository {
	return &ClientRepository{
		db: db,
	}
}

func (r *ClientRepository) Create(c *entity.Client) error {
	m := mapper.ClientToModel(c)
	return r.db.Create(m).Error
}

func (r *ClientRepository) GetByID(id, userID uint) (*entity.Client, error) {
	var m model.Client
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return mapper.ClientFromModel(&m), nil
}

func (r *ClientRepository) ListByUser(userID uint, page int, pageSize int, search string) ([]entity.Client, int64, error) {
	offset := (page - 1) * pageSize
	cond := "user_id = ?"
	args := []interface{}{userID}
	if search != "" {
		cond += " AND name LIKE @search"
		args = append(args, map[string]interface{}{"search": "%" + search + "%"})
	}

	var total int64
	if err := r.db.Model(&model.Client{}).
		Where(cond, args...).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []model.Client
	if err := r.db.Where(cond, args...).
		Limit(pageSize).
		Offset(offset).
		Order("id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	out := make([]entity.Client, 0, len(rows))
	for i := range rows {
		if e := mapper.ClientFromModel(&rows[i]); e != nil {
			out = append(out, *e)
		}
	}

	return out, total, nil
}

func (r *ClientRepository) Update(update entity.Client) error {
	updates := map[string]any{
		"name":    update.Name,
		"email":   update.Email,
		"phone":   update.Phone,
		"address": update.Address,
	}
	res := r.db.Model(&model.Client{}).
		Where("id = ? AND user_id = ?", update.ID, update.UserID).
		Updates(updates)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ClientRepository) Delete(id, userID uint) error {
	res := r.db.Unscoped().
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&model.Client{})

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *ClientRepository) SoftDeleteByUserID(userID uint) error {
	res := r.db.Where("user_id = ?", userID).
		Delete(&model.Client{})

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *ClientRepository) RestoreByUserID(userID uint) error {
	res := r.db.Unscoped().Model(&model.Client{}).
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Update("deleted_at", nil)

	if res.Error != nil {
		return res.Error
	}

	return nil
}
