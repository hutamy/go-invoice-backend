package postgres

import (
	"errors"

	"github.com/hutamy/go-invoice-backend/internal/adapter/mapper"
	pmodel "github.com/hutamy/go-invoice-backend/internal/adapter/repository/postgres/model"
	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
	"gorm.io/gorm"
)

type InvoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) ports.InvoiceRepository {
	return &InvoiceRepository{
		db: db,
	}
}

func (r *InvoiceRepository) Create(inv *entity.Invoice) error {
	m := mapper.InvoiceToModel(inv)
	return r.db.Create(m).Error
}

func (r *InvoiceRepository) GetByID(id, userID uint) (*entity.Invoice, error) {
	var m pmodel.Invoice
	err := r.db.Where("id = ? AND user_id = ?", id, userID).
		Preload("Items").
		First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if m.ClientID != nil {
		var client pmodel.Client
		err = r.db.Where("id = ?", *m.ClientID).First(&client).Error
		if err != nil {
			return nil, err
		}
		m.Client = &client
	}

	return mapper.InvoiceFromModel(&m), nil
}

func (r *InvoiceRepository) ListByUser(userID uint, page int, pageSize int, status string) ([]entity.Invoice, int64, error) {
	offset := (page - 1) * pageSize

	cond := "user_id = ?"
	args := []interface{}{userID}
	if status != "" {
		cond += " AND status = ?"
		args = append(args, status)
	}

	var total int64
	if err := r.db.Model(&pmodel.Invoice{}).
		Where(cond, args...).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []pmodel.Invoice
	if err := r.db.Where(cond, args...).
		Order("id DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	out := make([]entity.Invoice, 0, len(rows))
	for _, inv := range rows {
		if inv.ClientID != nil {
			var client pmodel.Client
			err := r.db.Where("id = ?", *inv.ClientID).First(&client).Error
			if err != nil {
				return nil, 0, err
			}

			inv.Client = &client
		}

		e := mapper.InvoiceFromModel(&inv)
		if e != nil {
			out = append(out, *e)
		}
	}

	return out, total, nil
}

func (r *InvoiceRepository) Update(update entity.Invoice) error {
	m := mapper.InvoiceToModel(&update)

	if err := r.db.Unscoped().
		Where("invoice_id = ?", m.ID).
		Delete(&pmodel.InvoiceItem{}).Error; err != nil {
		return err
	}

	for _, item := range update.Items {
		item.InvoiceID = m.ID
		if err := r.db.Create(&item).Error; err != nil {
			return err
		}
	}

	res := r.db.Model(&pmodel.Invoice{}).
		Where("id = ?", m.ID).
		Updates(m)

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *InvoiceRepository) Delete(id, userID uint) error {
	if err := r.db.Unscoped().
		Where("invoice_id = ?", id).
		Delete(&pmodel.InvoiceItem{}).Error; err != nil {
		return err
	}

	res := r.db.Unscoped().
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&pmodel.Invoice{})

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *InvoiceRepository) SoftDeleteByUserID(userID uint) error {
	res := r.db.Where("user_id = ?", userID).
		Delete(&pmodel.Invoice{})

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *InvoiceRepository) RestoreByUserID(userID uint) error {
	res := r.db.Unscoped().Model(&pmodel.Invoice{}).
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Update("deleted_at", nil)

	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *InvoiceRepository) UpdateStatus(id uint, userID uint, status entity.InvoiceStatus) error {
	res := r.db.Model(&pmodel.Invoice{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{"status": status})

	if res.Error != nil {
		return res.Error
	}

	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

func (r *InvoiceRepository) Summary(userID uint, status string) (float64, error) {
	var total float64
	cond := "user_id = ?"
	args := []interface{}{userID}
	if status != "" {
		cond += " AND status = ?"
		args = append(args, status)
	}

	if err := r.db.Model(&pmodel.Invoice{}).
		Where(cond, args...).
		Select("COALESCE(SUM(total), 0) as total").
		Scan(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
