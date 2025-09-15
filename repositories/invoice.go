package repositories

import (
	"fmt"
	"time"

	"github.com/hutamy/go-invoice-backend/dto"
	"github.com/hutamy/go-invoice-backend/models"
	"github.com/hutamy/go-invoice-backend/utils/errors"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	CreateInvoice(invoice *models.Invoice) error
	GetInvoiceByID(id uint) (*models.Invoice, error)
	ListInvoiceByUserID(userID uint) ([]models.Invoice, error)
	ListInvoiceByUserIDWithPagination(req dto.GetInvoicesRequest) ([]models.Invoice, int64, error)
	UpdateInvoice(id uint, req *dto.UpdateInvoiceRequest) error
	DeleteInvoice(id uint) error
	UpdateInvoiceStatus(id uint, status string) error
	InvoiceSummary(userID uint) (dto.SummaryInvoice, error)
	SoftDeleteAllByUserID(userID uint) error
	RestoreAllByUserID(userID uint) error
}

type invoiceRepository struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) InvoiceRepository {
	return &invoiceRepository{db: db}
}

func (r *invoiceRepository) CreateInvoice(invoice *models.Invoice) error {
	return r.db.Create(invoice).Error
}

func (r *invoiceRepository) GetInvoiceByID(id uint) (*models.Invoice, error) {
	var invoice models.Invoice
	if err := r.db.Preload("Items").First(&invoice, id).Error; err != nil {
		return nil, err
	}

	// Populate client information if client_id is not null
	if invoice.ClientID != nil {
		var client models.Client
		if err := r.db.First(&client, *invoice.ClientID).Error; err == nil {
			invoice.ClientName = &client.Name
			invoice.ClientEmail = &client.Email
			invoice.ClientPhone = &client.Phone
			invoice.ClientAddress = &client.Address
		}
	}

	return &invoice, nil
}

func (r *invoiceRepository) ListInvoiceByUserID(userID uint) ([]models.Invoice, error) {
	var invoices []models.Invoice
	if err := r.db.Where("user_id = ?", userID).Preload("Items").Order("created_at DESC").Find(&invoices).Error; err != nil {
		return nil, err
	}

	// Populate client information for invoices with client_id
	for i := range invoices {
		if invoices[i].ClientID != nil {
			var client models.Client
			if err := r.db.First(&client, *invoices[i].ClientID).Error; err == nil {
				invoices[i].ClientName = &client.Name
				invoices[i].ClientEmail = &client.Email
				invoices[i].ClientPhone = &client.Phone
				invoices[i].ClientAddress = &client.Address
			}
		}
	}

	return invoices, nil
}

func (r *invoiceRepository) ListInvoiceByUserIDWithPagination(req dto.GetInvoicesRequest) ([]models.Invoice, int64, error) {
	var invoices []models.Invoice
	var totalItems int64

	query := r.db.Where("user_id = ?", req.UserID)

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.Search != "" {
		searchTerm := "%" + req.Search + "%"
		query = query.Where(`
			invoice_number ILIKE ? OR 
			client_name ILIKE ? OR 
			client_email ILIKE ? OR 
			notes ILIKE ? OR 
			EXISTS (
				SELECT 1 FROM clients c 
				WHERE c.id = invoices.client_id 
				AND (c.name ILIKE ? OR c.email ILIKE ? OR c.phone ILIKE ? OR c.address ILIKE ?)
			)`,
			searchTerm, searchTerm, searchTerm, searchTerm,
			searchTerm, searchTerm, searchTerm, searchTerm)
	}

	if err := query.Model(&models.Invoice{}).Count(&totalItems).Error; err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.PageSize
	err := query.Preload("Items").
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&invoices).Error

	if err != nil {
		return nil, 0, err
	}

	// Populate client information for invoices with client_id
	for i := range invoices {
		if invoices[i].ClientID != nil {
			var client models.Client
			if err := r.db.First(&client, *invoices[i].ClientID).Error; err == nil {
				invoices[i].ClientName = &client.Name
				invoices[i].ClientEmail = &client.Email
				invoices[i].ClientPhone = &client.Phone
				invoices[i].ClientAddress = &client.Address
			}
		}
	}

	return invoices, totalItems, nil
}

func (r *invoiceRepository) UpdateInvoice(id uint, req *dto.UpdateInvoiceRequest) error {
	var invoice models.Invoice
	if err := r.db.Preload("Items").First(&invoice, id).Error; err != nil {
		return err
	}

	if req.ClientID != nil {
		invoice.ClientID = req.ClientID
		invoice.ClientName = nil
		invoice.ClientAddress = nil
		invoice.ClientEmail = nil
		invoice.ClientPhone = nil
	} else {
		invoice.ClientID = nil
		invoice.ClientName = req.ClientName
		invoice.ClientAddress = req.ClientAddress
		invoice.ClientEmail = req.ClientEmail
		invoice.ClientPhone = req.ClientPhone
	}

	if req.DueDate != nil {
		dueDate, err := time.Parse(time.DateOnly, *req.DueDate)
		if err != nil {
			return errors.ErrInvalidDateFormat
		}

		invoice.DueDate = dueDate
	}

	if req.IssueDate != nil {
		issueDate, err := time.Parse(time.DateOnly, *req.IssueDate)
		if err != nil {
			return errors.ErrInvalidDateFormat
		}

		invoice.IssueDate = issueDate
	}

	if req.Notes != nil {
		invoice.Notes = *req.Notes
	}

	if req.Status != nil {
		invoice.Status = *req.Status
	}

	if req.TaxRate != nil {
		invoice.TaxRate = *req.TaxRate
	}

	if req.InvoiceNumber != nil {
		invoice.InvoiceNumber = *req.InvoiceNumber
	}

	existingItems := map[uint]models.InvoiceItem{}
	for _, item := range invoice.Items {
		existingItems[item.ID] = item
	}

	var idsToKeep []uint
	var subtotal float64
	for _, itemReq := range req.Items {
		if itemReq.ID != nil {
			if existingItem, ok := existingItems[*itemReq.ID]; ok {
				existingItem.Description = itemReq.Description
				existingItem.Quantity = itemReq.Quantity
				existingItem.UnitPrice = itemReq.UnitPrice
				existingItem.Total = float64(itemReq.Quantity) * itemReq.UnitPrice
				subtotal += existingItem.Total
				if err := r.db.Save(&existingItem).Error; err != nil {
					return err
				}
				idsToKeep = append(idsToKeep, *itemReq.ID)
			} else {
				return fmt.Errorf("invoice item with ID %d not found", *itemReq.ID)
			}
		} else {
			newItem := models.InvoiceItem{
				InvoiceID:   invoice.ID,
				Description: itemReq.Description,
				Quantity:    itemReq.Quantity,
				UnitPrice:   itemReq.UnitPrice,
			}
			newItem.Total = float64(itemReq.Quantity) * itemReq.UnitPrice
			subtotal += newItem.Total
			if err := r.db.Create(&newItem).Error; err != nil {
				return err
			}
			idsToKeep = append(idsToKeep, newItem.ID)
		}
	}

	for _, existingItem := range invoice.Items {
		found := false
		for _, id := range idsToKeep {
			if existingItem.ID == id {
				found = true
				break
			}
		}
		if !found {
			if err := r.db.Delete(&existingItem).Error; err != nil {
				return err
			}
		}
	}

	invoice.Subtotal = subtotal
	invoice.Tax = invoice.TaxRate * subtotal / 100
	invoice.Total = invoice.Subtotal + invoice.Tax
	return r.db.Save(&invoice).Error
}

func (r *invoiceRepository) DeleteInvoice(id uint) error {
	return r.db.Model(&models.Invoice{}).
		Where("id = ?", id).
		Update("deleted_at", gorm.DeletedAt{Time: r.db.NowFunc(), Valid: true}).Error
}

func (r *invoiceRepository) UpdateInvoiceStatus(id uint, status string) error {
	var invoice models.Invoice
	if err := r.db.First(&invoice, id).Error; err != nil {
		return err
	}

	invoice.Status = status
	return r.db.Save(&invoice).Error
}

func (r *invoiceRepository) InvoiceSummary(userID uint) (summary dto.SummaryInvoice, err error) {
	r.db.Model(&models.Invoice{}).
		Where("user_id = ? AND status = ?", userID, "paid").
		Select("SUM(total) as total").
		Scan(&summary.Paid)

	r.db.Model(&models.Invoice{}).
		Where("user_id = ?", userID).
		Select("SUM(total) as total").
		Scan(&summary.TotalRevenue)

	return summary, nil
}

func (r *invoiceRepository) SoftDeleteAllByUserID(userID uint) error {
	return r.db.Model(&models.Invoice{}).
		Where("user_id = ?", userID).
		Update("deleted_at", gorm.DeletedAt{Time: r.db.NowFunc(), Valid: true}).Error
}

func (r *invoiceRepository) RestoreAllByUserID(userID uint) error {
	return r.db.Unscoped().Model(&models.Invoice{}).
		Where("user_id = ? AND deleted_at IS NOT NULL", userID).
		Update("deleted_at", nil).Error
}
