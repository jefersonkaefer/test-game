package controller

import (
	"context"

	"github.com/google/uuid"

	"game/api/internal/application/dto"
	"game/api/internal/domain/service"
)

type ClientController struct {
	clientService *service.ClientService
}

func NewClientController(clientService *service.ClientService) *ClientController {
	return &ClientController{
		clientService: clientService,
	}
}

func (c *ClientController) Create(ctx context.Context, username, password string) (res dto.CreateClientResponse, err error) {
	clientID, err := c.clientService.Create(ctx, username, password)
	if err != nil {
		return
	}

	res = dto.CreateClientResponse{
		ID: clientID.String(),
	}
	return
}

func (c *ClientController) GetBalance(ctx context.Context, clientID string) (res dto.GetBalanceResponse, err error) {
	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		return
	}
	err = c.clientService.RefreshWallet(ctx, clientUUID)
	if err != nil {
		return
	}
	balance, err := c.clientService.GetBalance(ctx, clientUUID)
	if err != nil {
		return
	}

	res = dto.GetBalanceResponse{
		Balance: balance,
	}
	return
}
