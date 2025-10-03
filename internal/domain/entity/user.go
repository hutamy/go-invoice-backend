package entity

type User struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	Email             string `json:"email"`
	Password          string `json:"-"`
	Address           string `json:"address"`
	Phone             string `json:"phone"`
	BankName          string `json:"bank_name"`
	BankAccountName   string `json:"bank_account_name"`
	BankAccountNumber string `json:"bank_account_number"`
	IsDeleted         bool   `json:"-"`
}
