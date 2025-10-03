package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/hutamy/go-invoice-backend/internal/domain/entity"
	"github.com/hutamy/go-invoice-backend/internal/domain/ports"
	response "github.com/hutamy/go-invoice-backend/internal/transport/http/response"
	"github.com/hutamy/go-invoice-backend/pkg/utils"
	"github.com/labstack/echo/v4"
)

type ClientHandler struct {
	UseCase ports.ClientUseCase
}

func NewClientHandler(uc ports.ClientUseCase) *ClientHandler {
	return &ClientHandler{
		UseCase: uc,
	}
}

type clientRequest struct {
	Name    string `json:"name" validate:"required"`
	Email   string `json:"email" validate:"required,email"`
	Phone   string `json:"phone" validate:"required"`
	Address string `json:"address" validate:"required"`
}

// @Summary Create Client
// @Description  Create a new client
// @Tags Client
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param request body clientRequest true "Client Request"
// @Success 201 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/clients [post]
func (h *ClientHandler) CreateClient(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	var req clientRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	client := &entity.Client{
		UserID:  userID,
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
	}
	if err := h.UseCase.Create(client); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusCreated, "created", nil)
}

// @Summary Get All Clients
// @Description  Get all clients
// @Tags Client
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param page query int false "Page"
// @Param page_size query int false "Page Size"
// @Param search query string false "Search"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/clients [get]
func (h *ClientHandler) GetAllClients(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	page := utils.ParseIntDefault(c.QueryParam("page"), 1)
	size := utils.ParseIntDefault(c.QueryParam("page_size"), 10)
	search := c.QueryParam("search")
	items, total, err := h.UseCase.ListByUser(userID, page, size, search)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "ok", map[string]any{
		"data": items,
		"pagination": map[string]any{
			"total_items": total,
			"page":        page,
			"page_size":   size,
			"total_pages": int(math.Ceil(float64(total) / float64(size))),
		},
	})
}

// @Summary Get Client By ID
// @Description  Get client by id
// @Tags Client
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Client ID"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/clients/{id} [get]
func (h *ClientHandler) GetClientByID(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	clientID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || clientID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	client, err := h.UseCase.GetByID(uint(clientID), userID)
	if err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	if client == nil {
		return response.Response(c, http.StatusNotFound, "not found", nil)
	}

	return response.Response(c, http.StatusOK, "ok", client)
}

// @Summary Update Client
// @Description  Update client by id
// @Tags Client
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Client ID"
// @Param request body clientRequest true "Client Request"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/clients/{id} [put]
func (h *ClientHandler) UpdateClient(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	clientID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || clientID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	var req clientRequest
	if err := c.Bind(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, "invalid request", nil)
	}

	if err := c.Validate(&req); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	update := entity.Client{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		UserID:  userID,
		ID:      uint(clientID),
	}
	if err := h.UseCase.Update(update); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "updated", nil)
}

// @Summary Delete Client
// @Description  Delete client by id
// @Tags Client
// @Accept json
// @Produce json
// @Security     BearerAuth
// @Param id path int true "Client ID"
// @Success 200 {object} response.GenericResponse
// @Failure 400 {object} response.GenericResponse
// @Router /v1/protected/clients/{id} [delete]
func (h *ClientHandler) DeleteClient(c echo.Context) error {
	id := c.Get("user_id")
	userID, ok := id.(uint)
	if !ok || userID == 0 {
		return response.Response(c, http.StatusUnauthorized, "unauthorized", nil)
	}

	clientID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || clientID == 0 {
		return response.Response(c, http.StatusBadRequest, "invalid id", nil)
	}

	if err := h.UseCase.Delete(uint(clientID), userID); err != nil {
		return response.Response(c, http.StatusBadRequest, err.Error(), nil)
	}

	return response.Response(c, http.StatusOK, "deleted", nil)
}
