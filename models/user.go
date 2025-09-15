package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Name              string         `json:"name" gorm:"not null"`
	Email             string         `json:"email" gorm:"not null;uniqueIndex"`
	Password          string         `json:"-" gorm:"not null"`
	Address           string         `json:"address"`
	Phone             string         `json:"phone"`
	BankName          string         `json:"bank_name"`
	BankAccountName   string         `json:"bank_account_name"`
	BankAccountNumber string         `json:"bank_account_number"`
	CreatedAt         time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt         time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index" swaggerignore:"true"`
}
