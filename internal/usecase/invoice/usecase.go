package invoice

import (
	"context"
	"fmt"
	"strconv"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type UseCase struct {
	InvoiceRepo ports.InvoiceRepository
	ClientRepo  ports.ClientRepository
	AuthRepo    ports.AuthRepository
}

func NewUseCase(
	invRepo ports.InvoiceRepository,
	clientRepo ports.ClientRepository,
	authRepo ports.AuthRepository,
) ports.InvoiceUseCase {
	return &UseCase{
		InvoiceRepo: invRepo,
		ClientRepo:  clientRepo,
		AuthRepo:    authRepo,
	}
}

func (u *UseCase) Create(inv *entity.Invoice) error {
	return u.InvoiceRepo.Create(inv)
}

func (u *UseCase) GetByID(id, userID uint) (*entity.Invoice, error) {
	return u.InvoiceRepo.GetByID(id, userID)
}

func (u *UseCase) ListByUser(userID uint, page, pageSize int, status string) ([]entity.Invoice, int64, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 10
	}

	return u.InvoiceRepo.ListByUser(userID, page, pageSize, status)
}

func (u *UseCase) Update(update entity.Invoice) error {
	return u.InvoiceRepo.Update(update)
}

func (u *UseCase) Delete(id uint, userID uint) error {
	return u.InvoiceRepo.Delete(id, userID)
}

func (u *UseCase) UpdateStatus(id uint, userID uint, status entity.InvoiceStatus) error {
	return u.InvoiceRepo.UpdateStatus(id, userID, status)
}

func (u *UseCase) Summary(userID uint) (paid, revenue float64, err error) {
	paid, err = u.InvoiceRepo.Summary(userID, string(entity.InvoiceStatusPaid))
	if err != nil {
		return 0, 0, err
	}

	total, err := u.InvoiceRepo.Summary(userID, "")
	if err != nil {
		return 0, 0, err
	}

	return paid, total, nil
}

func (u *UseCase) GeneratePDF(id, userID uint) ([]byte, error) {
	invoice, err := u.InvoiceRepo.GetByID(id, userID)
	if err != nil {
		return nil, err
	}

	var client *entity.Client
	if invoice.ClientID != nil {
		client, err = u.ClientRepo.GetByID(*invoice.ClientID, userID)
		if err != nil {
			return nil, err
		}
	} else {
		client = &entity.Client{
			ID:      0,
			Name:    *invoice.ClientName,
			Email:   *invoice.ClientEmail,
			Address: *invoice.ClientAddress,
			Phone:   *invoice.ClientPhone,
		}
	}

	user, err := u.AuthRepo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}

	htmlContent := u.generateTemplate(*invoice, *user, *client)
	return u.generatePdf(htmlContent)
}

func (u *UseCase) GeneratePDFPublic(invoice *entity.Invoice) ([]byte, error) {
	for i, it := range invoice.Items {
		total := float64(it.Quantity) * it.UnitPrice
		invoice.Items[i].Total = total
		invoice.Subtotal += total
	}

	invoice.Tax = invoice.Subtotal * invoice.TaxRate
	invoice.Total = invoice.Subtotal + invoice.Tax + invoice.DeliveryFee

	htmlContent := u.generateTemplate(*invoice, invoice.User, invoice.Client)
	return u.generatePdf(htmlContent)
}

func (u *UseCase) generatePdf(htmlContent string) ([]byte, error) {
	// Setup headless browser
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var pdfBuf []byte
	err := chromedp.Run(ctx,
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(`document.documentElement.innerHTML = `+strconv.Quote(htmlContent), nil).Do(ctx)
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdfBuf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, err
	}

	return pdfBuf, nil
}

func (u *UseCase) generateTemplate(invoice entity.Invoice, user entity.User, client entity.Client) string {
	p := message.NewPrinter(language.English)
	items := ""
	for _, it := range invoice.Items {
		items += fmt.Sprintf(`
			<tr>
				<td>%s</td>
				<td>%d</td>
				<td>IDR %s</td>
				<td>IDR %s</td>
			</tr>`, it.Description, it.Quantity, p.Sprintf("%.2f", it.UnitPrice), p.Sprintf("%.2f", it.Total))
	}

	deliveryRate := ""
	if invoice.DeliveryFee > 0 {
		deliveryRate = fmt.Sprintf(`
			<div class="invoice-tax">
				<span>Delivery Fee:</span>
				<span>IDR %s</span>
			</div>
		`, p.Sprintf("%.2f", invoice.DeliveryFee))
	}

	terms := ""
	if invoice.Notes != "" {
		terms = fmt.Sprintf("<div><span style=\"font-weight: 500;\">Terms:</span> %s</div>", invoice.Notes)
	}

	template := fmt.Sprintf(`
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="utf-8" />
		<title>Invoice %s</title>
		<meta name="viewport" content="width=device-width, initial-scale=1.0" />
		<style>
		:root {
			--primary-color: #111827;
			--text-color: #1f2937;
			--light-gray: #f9fafb;
			--border-color: #e5e7eb;
		}

		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		body {
			font-family: "Inter", "Segoe UI", sans-serif;
			color: var(--text-color);
			line-height: 1.5;
			background-color: white;
			padding: 32px 20px;
		}

		.invoice-container {
			max-width: 800px;
			margin: 0 auto;
			background: white;
			padding: 32px;
		}

		.invoice-header {
			display: flex;
			justify-content: space-between;
			align-items: flex-start;
			margin-bottom: 48px;
		}

		.invoice-title {
			font-weight: 700;
			font-size: 36px;
			color: #111827;
			margin-bottom: 8px;
		}

		.invoice-id {
			font-size: 14px;
			color: #6b7280;
		}

		.invoice-dates {
			text-align: right;
			font-size: 14px;
			color: #4b5563;
			line-height: 1.6;
		}

		.invoice-dates > div {
			margin-bottom: 4px;
		}

		.invoice-parties {
			display: grid;
			grid-template-columns: 1fr 1fr;
			margin-bottom: 48px;
			gap: 48px;
		}

		.invoice-parties h3 {
			font-size: 14px;
			font-weight: 600;
			text-transform: uppercase;
			letter-spacing: 0.025em;
			color: #4b5563;
			margin-bottom: 16px;
		}

		.party-info {
			font-size: 14px;
			line-height: 1.6;
			color: #1f2937;
		}

		.invoice-table {
			width: 100%%;
			border-collapse: collapse;
			margin-bottom: 32px;
		}

		.invoice-table th {
			padding: 12px 8px;
			text-align: left;
			background-color: #f9fafb;
			font-weight: 600;
			font-size: 14px;
			border-bottom: 2px solid #e5e7eb;
		}

		.invoice-table td {
			padding: 16px 8px;
			font-size: 14px;
			color: #1f2937;
			border-bottom: 1px solid #f3f4f6;
		}

		.invoice-table tr:last-child td {
			border-bottom: none;
		}

		.invoice-table th:last-child,
		.invoice-table td:last-child {
			text-align: right;
		}

		.invoice-totals {
			display: flex;
			flex-direction: column;
			align-items: flex-end;
			margin-bottom: 32px;
		}

		.invoice-subtotal,
		.invoice-tax {
			display: flex;
			justify-content: space-between;
			width: 320px;
			padding: 8px 0;
			font-size: 14px;
		}

		.invoice-subtotal span:first-child,
		.invoice-tax span:first-child {
			color: #4b5563;
		}

		.invoice-subtotal span:last-child,
		.invoice-tax span:last-child {
			color: #1f2937;
		}

		.invoice-total {
			display: flex;
			justify-content: space-between;
			width: 320px;
			padding: 12px 0;
			border-top: 1px solid #d1d5db;
		}

		.invoice-total-label {
			font-size: 16px;
			font-weight: 600;
			color: #111827;
		}

		.invoice-total-amount {
			font-size: 16px;
			font-weight: 700;
			color: #111827;
		}

		.invoice-notes {
			margin-bottom: 32px;
			font-size: 14px;
			color: #374151;
			line-height: 1.6;
		}

		.invoice-notes > div {
			margin-bottom: 16px;
		}

		.bank-details {
			padding: 24px;
			background-color: #f3f4f6;
			border-radius: 4px;
			font-size: 14px;
		}

		.bank-details h4 {
			font-size: 14px;
			font-weight: 600;
			text-transform: uppercase;
			letter-spacing: 0.025em;
			color: #4b5563;
			margin-bottom: 16px;
		}

		.bank-details-grid {
			color: #374151;
		}

		.bank-details-grid > div {
			margin-bottom: 8px;
		}

		.bank-details-label {
			font-weight: 500;
			display: inline-block;
			min-width: 150px;
		}

		@media (max-width: 768px) {
			.invoice-header,
			.invoice-parties {
			flex-direction: column;
			}

			.invoice-dates,
			.invoice-parties div:last-child {
			margin-top: 20px;
			text-align: left;
			}
		}
		</style>
	</head>
	<body>
		<div class="invoice-container">
		<div class="invoice-header">
			<div>
			<div class="invoice-title">INVOICE</div>
			<div class="invoice-id">%s</div>
			</div>
			<div class="invoice-dates">
			<div>Issue Date: %s</div>
			<div>Due Date: %s</div>
			</div>
		</div>

		<div class="invoice-parties">
			<div>
				<h3>From</h3>
				<div class="party-info">
					%s<br />
					%s <br />
					%s<br />
					%s
				</div>
				</div>
				<div>
				<h3>To</h3>
				<div class="party-info">
					%s <br />
					%s<br />
					%s<br />
					%s
				</div>
			</div>
		</div>

		<table class="invoice-table">
			<thead>
			<tr>
				<th>Description</th>
				<th>Quantity</th>
				<th>Unit Price</th>
				<th>Total</th>
			</tr>
			</thead>
			<tbody>
			%s
			</tbody>
		</table>

		<div class="invoice-totals">
			<div class="invoice-subtotal">
				<span>Subtotal:</span>
				<span>IDR %s</span>
			</div>
			<div class="invoice-tax">
				<span>Tax (%s%%):</span>
				<span>IDR %s</span>	
			</div>
			%s
			<div class="invoice-total">
				<span class="invoice-total-label">Total:</span>
				<span class="invoice-total-amount"
					>IDR %s</span
				>
			</div>
		</div>

		<div class="invoice-notes">
			%s
			<div><strong>Thank you</strong> for your business!</div>
		</div>

		<div class="bank-details">
			<h4>Bank Account Details</h4>
			<div class="bank-details-grid">
				<div>
					<span class="bank-details-label">Bank Name:</span>
					<span>%s</span>
				</div>
				<div>
					<span class="bank-details-label">Account Name:</span>
					<span>%s</span>
				</div>
				<div>
					<span class="bank-details-label">Account Number:</span>
					<span>%s</span>
				</div>
			</div>
		</div>
		</div>
	</body>
	</html>
	`, invoice.InvoiceNumber,
		invoice.InvoiceNumber,
		invoice.DueDate.Format("02 Jan 2006"),
		invoice.IssueDate.Format("02 Jan 2006"),
		user.Name,
		user.Address,
		user.Email,
		user.Phone,
		client.Name,
		client.Address,
		client.Email,
		client.Phone,
		items,
		p.Sprintf("%.2f", invoice.Subtotal),
		p.Sprintf("%.2f", invoice.TaxRate),
		p.Sprintf("%.2f", invoice.Tax),
		deliveryRate,
		p.Sprintf("%.2f", invoice.Total),
		terms,
		user.BankName,
		user.BankAccountName,
		user.BankAccountNumber,
	)

	return template
}
