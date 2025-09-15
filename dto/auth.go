package dto

type SignUpRequest struct {
	Name              string `json:"name" binding:"required"`
	Email             string `json:"email" binding:"required,email"`
	Password          string `json:"password" binding:"required,min=6"`
	Address           string `json:"address" binding:"required"`
	Phone             string `json:"phone" binding:"required"`
	BankName          string `json:"bank_name" binding:"required"`
	BankAccountName   string `json:"bank_account_name" binding:"required"`
	BankAccountNumber string `json:"bank_account_number" binding:"required,numeric,gt=0"`
}

type SignInRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type UpdateUserProfileRequest struct {
	Name    *string `json:"name"`
	Email   *string `json:"email" validate:"omitempty,email"` // Validate email format
	Address *string `json:"address"`
	Phone   *string `json:"phone"`
	UserID  uint    `json:"-"` // This field is used internally to identify the user being updated
}

type UpdateUserBankingRequest struct {
	BankName          *string `json:"bank_name"`
	BankAccountName   *string `json:"bank_account_name"`
	BankAccountNumber *string `json:"bank_account_number" validate:"omitempty,numeric,gt=0"` // Validate bank account number format (numeric and > 0)
	UserID            uint    `json:"-"`                                                     // This field is used internally to identify the user being updated
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
	UserID      uint   `json:"-"` // This field is used internally to identify the user being updated
}
