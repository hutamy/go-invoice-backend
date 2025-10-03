package handlers

import (
	"net/http"

	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
	response "github.com/hutamy/go-invoice-backend/internal/transport/http/response"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	UseCase ports.AuthUseCase
}

func NewAuthHandler(uc ports.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		UseCase: uc,
	}
}

func (h *AuthHandler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}

type signUpRequest struct {
	Name              string `json:"name" validate:"required"`
	Email             string `json:"email" validate:"required,email"`
	Password          string `json:"password" validate:"required,min=6"`
	Address           string `json:"address" binding:"required"`
	Phone             string `json:"phone" binding:"required"`
	BankName          string `json:"bank_name" binding:"required"`
	BankAccountName   string `json:"bank_account_name" binding:"required"`
	BankAccountNumber string `json:"bank_account_number" binding:"required,numeric,gt=0"`
}

type signInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type updateProfileRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Address string `json:"address" validate:"required"`
	Phone   string `json:"phone" validate:"required"`
}

type updateBankingRequest struct {
	BankName          string `json:"bank_name" binding:"required"`
	BankAccountName   string `json:"bank_account_name" binding:"required"`
	BankAccountNumber string `json:"bank_account_number" binding:"required,numeric,gt=0"`
}

type updatePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type refreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// @Summary Sign Up
// @Description  Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body signUpRequest true "Sign Up Request"
// @Success 201 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/public/auth/sign-up [post]
func (h *AuthHandler) SignUp(c echo.Context) error {
	var req signUpRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	user := &entity.User{
		Name:              req.Name,
		Email:             req.Email,
		Password:          req.Password,
		Address:           req.Address,
		Phone:             req.Phone,
		BankName:          req.BankName,
		BankAccountName:   req.BankAccountName,
		BankAccountNumber: req.BankAccountNumber,
	}

	access, refresh, err := h.UseCase.SignUp(user)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusCreated, "sign up success", map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// @Summary Sign In
// @Description  Sign in with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body signInRequest true "Sign In Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/public/auth/sign-in [post]
func (h *AuthHandler) SignIn(c echo.Context) error {
	var req signInRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	access, refresh, err := h.UseCase.SignIn(req.Email, req.Password)
	if err != nil {
		return response.Response(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "sign in success", map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// @Summary Refresh Token
// @Description  Refresh access token with refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body refreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c echo.Context) error {
	var req refreshTokenRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	access, refresh, err := h.UseCase.RefreshToken(req.RefreshToken)
	if err != nil {
		return response.Response(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "refresh token success", map[string]string{
		"access_token":  access,
		"refresh_token": refresh,
	})
}

// @Summary Me
// @Description  Get current user
// @Tags Auth
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/me [get]
func (h *AuthHandler) Me(c echo.Context) error {
	id := c.Get("user_id")
	user_id, ok := id.(uint)
	if !ok {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	user, err := h.UseCase.Me(user_id)
	if err != nil {
		return response.Response(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", user)
}

// @Summary Update Profile
// @Description  Update current user profile
// @Tags Auth
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param request body updateProfileRequest true "Update Profile Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/me/profile [put]
func (h *AuthHandler) UpdateProfile(c echo.Context) error {
	id := c.Get("user_id")
	user_id, ok := id.(uint)
	if !ok {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	var req updateProfileRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	user := entity.User{
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Phone:   req.Phone,
	}
	if err := h.UseCase.UpdateUserProfile(user_id, user); err != nil {
		return response.Response(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", nil)
}

// @Summary Update Banking
// @Description  Update current user banking
// @Tags Auth
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param request body updateBankingRequest true "Update Banking Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/me/banking [put]
func (h *AuthHandler) UpdateBanking(c echo.Context) error {
	id := c.Get("user_id")
	user_id, ok := id.(uint)
	if !ok {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	var req updateBankingRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	user := entity.User{
		BankName:          req.BankName,
		BankAccountName:   req.BankAccountName,
		BankAccountNumber: req.BankAccountNumber,
	}
	if err := h.UseCase.UpdateUserBanking(user_id, user); err != nil {
		return response.Response(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", nil)
}

// @Summary Change Password
// @Description  Change current user password
// @Tags Auth
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param request body updatePasswordRequest true "Update Password Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/me/change-password [post]
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	id := c.Get("user_id")
	user_id, ok := id.(uint)
	if !ok {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	var req updatePasswordRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	if err := h.UseCase.ChangePassword(user_id, req.OldPassword, req.NewPassword); err != nil {
		return response.Response(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", nil)
}

// @Summary Deactivate User
// @Description  Deactivate current user
// @Tags Auth
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/me/deactivate [post]
func (h *AuthHandler) DeactivateUser(c echo.Context) error {
	id := c.Get("user_id")
	user_id, ok := id.(uint)
	if !ok {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	if err := h.UseCase.DeactivateUser(user_id); err != nil {
		return response.Response(c, http.StatusUnauthorized, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", nil)
}
