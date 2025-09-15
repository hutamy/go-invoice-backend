package errors

import e "errors"

var (
	ErrUserAlreadyExists   = e.New("email already exists")
	ErrLoginFailed         = e.New("invalid email and password")
	ErrAccountDeactivated  = e.New("account has been deactivated")
	ErrBadRequest          = e.New("please check your input")
	ErrFailedGenerateToken = e.New("failed to generate token")
	ErrUserNotFound        = e.New("user not found")
	ErrInvalidToken        = e.New("invalid token")
	ErrUnauthorized        = e.New("unauthorized access")
	ErrNotFound            = e.New("resource not found")
	ErrInvalidDateFormat   = e.New("invalid date format, expected YYYY-MM-DD")
	ErrInvalidOldPassword  = e.New("invalid old password")
)
