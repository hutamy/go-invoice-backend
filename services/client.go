package services

import (
	"github.com/hutamy/go-invoice-backend/dto"
	"github.com/hutamy/go-invoice-backend/models"
	"github.com/hutamy/go-invoice-backend/repositories"
	"github.com/hutamy/go-invoice-backend/utils"
)

type ClientService interface {
	CreateClient(req dto.CreateClientRequest) error
	GetAllClientsByUserID(userID uint) ([]models.Client, error)
	GetAllClientsByUserIDWithPagination(req dto.GetClientsRequest) (utils.PaginatedResponse, error)
	GetClientByID(id, userID uint) (*models.Client, error)
	UpdateClient(req dto.UpdateClientRequest) error
	DeleteClient(id, userID uint) error
}

type clientService struct {
	clientRepo repositories.ClientRepository
}

func NewClientService(clientRepo repositories.ClientRepository) ClientService {
	return &clientService{clientRepo: clientRepo}
}

func (s *clientService) CreateClient(req dto.CreateClientRequest) error {
	client := &models.Client{
		Name:    req.Name,
		Email:   req.Email,
		Phone:   req.Phone,
		Address: req.Address,
		UserID:  req.UserID,
	}
	return s.clientRepo.CreateClient(client)
}

func (s *clientService) GetAllClientsByUserID(userID uint) ([]models.Client, error) {
	return s.clientRepo.GetAllByUserID(userID)
}

func (s *clientService) GetAllClientsByUserIDWithPagination(req dto.GetClientsRequest) (utils.PaginatedResponse, error) {
	// Set default values for pagination
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	clients, totalItems, err := s.clientRepo.GetAllByUserIDWithPagination(req)
	if err != nil {
		return utils.PaginatedResponse{}, err
	}

	pagination := utils.CalculatePagination(req.Page, req.PageSize, totalItems)
	return utils.PaginatedData(clients, pagination), nil
}

func (s *clientService) GetClientByID(id, userID uint) (*models.Client, error) {
	return s.clientRepo.GetClientByID(id, userID)
}

func (s *clientService) UpdateClient(req dto.UpdateClientRequest) error {
	client, err := s.GetClientByID(req.ID, req.UserID)
	if err != nil {
		return err
	}

	if req.Name != nil {
		client.Name = *req.Name
	}

	if req.Email != nil {
		client.Email = *req.Email
	}

	if req.Address != nil {
		client.Address = *req.Address
	}

	if req.Phone != nil {
		client.Phone = *req.Phone
	}

	return s.clientRepo.UpdateClient(client)
}

func (s *clientService) DeleteClient(id, userID uint) error {
	return s.clientRepo.DeleteClient(id, userID)
}
