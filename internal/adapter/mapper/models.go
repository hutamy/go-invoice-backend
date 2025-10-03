package mapper

import (
	pmodel "github.com/hutamy/go-invoice-backend/internal/adapter/repository/postgres/model"
	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
)

func UserToModel(u *entity.User) *pmodel.User {
	if u == nil {
		return nil
	}

	return &pmodel.User{
		ID:                u.ID,
		Name:              u.Name,
		Email:             u.Email,
		Password:          u.Password,
		Address:           u.Address,
		Phone:             u.Phone,
		BankName:          u.BankName,
		BankAccountName:   u.BankAccountName,
		BankAccountNumber: u.BankAccountNumber,
	}
}

func UserFromModel(m *pmodel.User) *entity.User {
	if m == nil {
		return nil
	}

	return &entity.User{
		ID:                m.ID,
		Name:              m.Name,
		Email:             m.Email,
		Password:          m.Password,
		Address:           m.Address,
		Phone:             m.Phone,
		BankName:          m.BankName,
		BankAccountName:   m.BankAccountName,
		BankAccountNumber: m.BankAccountNumber,
		IsDeleted:         m.DeletedAt.Valid,
	}
}

func ClientToModel(c *entity.Client) *pmodel.Client {
	if c == nil {
		return nil
	}

	return &pmodel.Client{
		ID:      c.ID,
		UserID:  c.UserID,
		Name:    c.Name,
		Email:   c.Email,
		Address: c.Address,
		Phone:   c.Phone,
	}
}

func ClientFromModel(m *pmodel.Client) *entity.Client {
	if m == nil {
		return nil
	}

	return &entity.Client{
		ID:      m.ID,
		UserID:  m.UserID,
		Name:    m.Name,
		Email:   m.Email,
		Address: m.Address,
		Phone:   m.Phone,
	}
}

func InvoiceItemToModel(it *entity.InvoiceItem) *pmodel.InvoiceItem {
	if it == nil {
		return nil
	}

	return &pmodel.InvoiceItem{
		ID:          it.ID,
		InvoiceID:   it.InvoiceID,
		Description: it.Description,
		Quantity:    it.Quantity,
		UnitPrice:   it.UnitPrice,
		Total:       float64(it.Quantity) * it.UnitPrice,
	}
}

func InvoiceItemFromModel(m *pmodel.InvoiceItem) *entity.InvoiceItem {
	if m == nil {
		return nil
	}
	return &entity.InvoiceItem{
		ID:          m.ID,
		InvoiceID:   m.InvoiceID,
		Description: m.Description,
		Quantity:    m.Quantity,
		UnitPrice:   m.UnitPrice,
	}
}

func InvoiceToModel(inv *entity.Invoice) *pmodel.Invoice {
	if inv == nil {
		return nil
	}

	m := &pmodel.Invoice{
		ID:            inv.ID,
		UserID:        inv.UserID,
		ClientID:      inv.ClientID,
		ClientName:    inv.ClientName,
		ClientEmail:   inv.ClientEmail,
		ClientAddress: inv.ClientAddress,
		ClientPhone:   inv.ClientPhone,
		InvoiceNumber: inv.InvoiceNumber,
		IssueDate:     inv.IssueDate,
		DueDate:       inv.DueDate,
		Status:        string(inv.Status),
		Notes:         inv.Notes,
		TaxRate:       inv.TaxRate,
		DeliveryFee:   inv.DeliveryFee,
	}

	m.Items = make([]pmodel.InvoiceItem, 0, len(inv.Items))
	for i := range inv.Items {
		if mm := InvoiceItemToModel(&inv.Items[i]); mm != nil {
			m.Items = append(m.Items, *mm)
			m.Subtotal += mm.Total
		}
	}

	m.Tax = m.Subtotal * m.TaxRate / 100
	m.Total = m.Subtotal + m.Tax + m.DeliveryFee
	return m
}

func InvoiceFromModel(m *pmodel.Invoice) *entity.Invoice {
	if m == nil {
		return nil
	}

	inv := &entity.Invoice{
		ID:            m.ID,
		UserID:        m.UserID,
		ClientID:      m.ClientID,
		ClientName:    m.ClientName,
		ClientEmail:   m.ClientEmail,
		ClientAddress: m.ClientAddress,
		ClientPhone:   m.ClientPhone,
		InvoiceNumber: m.InvoiceNumber,
		IssueDate:     m.IssueDate,
		DueDate:       m.DueDate,
		Status:        string(m.Status),
		Notes:         m.Notes,
		Subtotal:      m.Subtotal,
		Tax:           m.Tax,
		TaxRate:       m.TaxRate,
		DeliveryFee:   m.DeliveryFee,
		Total:         m.Total,
	}

	if m.ClientID != nil {
		inv.ClientName = &m.Client.Name
		inv.ClientEmail = &m.Client.Email
		inv.ClientAddress = &m.Client.Address
		inv.ClientPhone = &m.Client.Phone
	}

	inv.Items = make([]entity.InvoiceItem, 0, len(m.Items))
	for i := range m.Items {
		inv.Items = append(inv.Items, *InvoiceItemFromModel(&m.Items[i]))
	}

	return inv
}
