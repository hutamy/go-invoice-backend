package controllers

import (
	"net/http"
	"strconv"

	"github.com/hutamy/go-invoice-backend/dto"
	"github.com/hutamy/go-invoice-backend/services"
	"github.com/hutamy/go-invoice-backend/utils"
	"github.com/hutamy/go-invoice-backend/utils/errors"
	"github.com/labstack/echo/v4"
)

type ClientController struct {
	clientService services.ClientService
}

func NewClientController(clientService services.ClientService) *ClientController {
	return &ClientController{clientService: clientService}
}

// @Summary      Create a new client
// @Description  Creates a new client for the authenticated user
// @Tags         clients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        client  body      dto.CreateClientRequest  true  "Client data"
// @Success      201     {object}  utils.GenericResponse
// @Failure      400     {object}  utils.GenericResponse
// @Failure      500     {object}  utils.GenericResponse
// @Router       /v1/protected/clients [post]
func (c *ClientController) CreateClient(ctx echo.Context) error {
	userID := ctx.Get("user_id").(uint)

	var client dto.CreateClientRequest
	if err := ctx.Bind(&client); err != nil {
		return utils.Response(ctx, http.StatusBadRequest, errors.ErrBadRequest.Error(), nil)
	}

	client.UserID = userID
	if err := c.clientService.CreateClient(client); err != nil {
		return utils.Response(ctx, http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.Response(ctx, http.StatusCreated, "Client created successfully", nil)
}

// @Summary      Get all clients
// @Description  Retrieves all clients for the authenticated user with pagination (default: page=1, page_size=10)
// @Tags         clients
// @Produce      json
// @Security     BearerAuth
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        page_size query     int     false  "Page size (default: 10, max: 100)"
// @Param        search    query     string  false  "Search term for filtering clients"
// @Param        all       query     bool    false  "Return all clients without pagination (use with caution)"
// @Success      200  {object}  utils.GenericResponse
// @Failure      400  {object}  utils.GenericResponse
// @Failure      500  {object}  utils.GenericResponse
// @Router       /v1/protected/clients [get]
func (c *ClientController) GetAllClients(ctx echo.Context) error {
	userID := ctx.Get("user_id").(uint)

	// Check if user explicitly wants all clients without pagination
	all := ctx.QueryParam("all") == "true"
	if all {
		// Use non-paginated response (backward compatibility)
		clients, err := c.clientService.GetAllClientsByUserID(userID)
		if err != nil {
			return utils.Response(ctx, http.StatusInternalServerError, err.Error(), nil)
		}
		return utils.Response(ctx, http.StatusOK, "All clients retrieved successfully", clients)
	}

	// Always use pagination by default
	var paginationReq dto.PaginationRequest
	ctx.Bind(&paginationReq) // Bind query parameters, ignore errors

	req := dto.GetClientsRequest{
		UserID:            userID,
		PaginationRequest: paginationReq,
	}

	paginatedClients, err := c.clientService.GetAllClientsByUserIDWithPagination(req)
	if err != nil {
		return utils.Response(ctx, http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.Response(ctx, http.StatusOK, "Clients retrieved successfully", paginatedClients)
}

// @Summary      Get client by ID
// @Description  Retrieves a client by its ID for the authenticated user
// @Tags         clients
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Client ID"
// @Success      200  {object}  utils.GenericResponse
// @Failure      400  {object}  utils.GenericResponse
// @Failure      404  {object}  utils.GenericResponse
// @Failure      500  {object}  utils.GenericResponse
// @Router       /v1/protected/clients/{id} [get]
func (c *ClientController) GetClientByID(ctx echo.Context) error {
	userID := ctx.Get("user_id").(uint)
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return utils.Response(ctx, http.StatusBadRequest, errors.ErrBadRequest.Error(), nil)
	}

	client, err := c.clientService.GetClientByID(uint(id), userID)
	if err != nil {
		if err == errors.ErrNotFound {
			return utils.Response(ctx, http.StatusNotFound, err.Error(), nil)
		}

		return utils.Response(ctx, http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.Response(ctx, http.StatusOK, "Client retrieved successfully", client)
}

// @Summary      Update client
// @Description  Updates a client by its ID for the authenticated user
// @Tags         clients
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path      int           true  "Client ID"
// @Param        client  body      dto.UpdateClientRequest true  "Client data"
// @Success      200     {object}  utils.GenericResponse
// @Failure      400     {object}  utils.GenericResponse
// @Failure      500     {object}  utils.GenericResponse
// @Router       /v1/protected/clients/{id} [put]
func (c *ClientController) UpdateClient(ctx echo.Context) error {
	var client dto.UpdateClientRequest
	if err := ctx.Bind(&client); err != nil {
		return utils.Response(ctx, http.StatusBadRequest, errors.ErrBadRequest.Error(), nil)
	}

	client.UserID = ctx.Get("user_id").(uint)
	if err := c.clientService.UpdateClient(client); err != nil {
		return utils.Response(ctx, http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.Response(ctx, http.StatusOK, "Client updated successfully", nil)
}

// @Summary      Delete client
// @Description  Soft deletes a client by its ID for the authenticated user (can be restored)
// @Tags         clients
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Client ID"
// @Success      200  {object}  utils.GenericResponse
// @Failure      400  {object}  utils.GenericResponse
// @Failure      404  {object}  utils.GenericResponse
// @Failure      500  {object}  utils.GenericResponse
// @Router       /v1/protected/clients/{id} [delete]
func (c *ClientController) DeleteClient(ctx echo.Context) error {
	userID := ctx.Get("user_id").(uint)
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		return utils.Response(ctx, http.StatusBadRequest, errors.ErrBadRequest.Error(), nil)
	}

	if err := c.clientService.DeleteClient(uint(id), userID); err != nil {
		if err == errors.ErrNotFound {
			return utils.Response(ctx, http.StatusNotFound, err.Error(), nil)
		}

		return utils.Response(ctx, http.StatusInternalServerError, err.Error(), nil)
	}

	return utils.Response(ctx, http.StatusOK, "Client deleted successfully", nil)
}
