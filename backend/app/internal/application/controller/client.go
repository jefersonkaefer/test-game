package controller

import (
	"context"

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
		ID:       clientID,
		Username: username,
	}
	return
}

func (c *ClientController) GetBalance(ctx context.Context, clientID string) (res dto.ClientDTO, err error) {
	client, err := c.clientService.GetByUsername(ctx, clientID)
	if err != nil {
		return
	}

	res = dto.ClientDTO{
		ID:       client.GetID().String(),
		Username: client.GetUsername(),
		Balance:  client.GetBalance(),
	}
	return
}

func (c *ClientController) UpdateBalance(ctx context.Context, clientID string, amount float64) (res dto.UpdateClientBalanceResponse, err error) {
	client, err := c.clientService.GetByUsername(ctx, clientID)
	if err != nil {
		return
	}

	if amount > 0 {
		client.Credit(amount)
	} else {
		client.Debit(-amount)
	}

	err = c.clientService.Update(ctx, client)
	if err != nil {
		return
	}

	res = dto.UpdateClientBalanceResponse{
		ID:      client.GetID().String(),
		Balance: client.GetBalance(),
	}
	return
}
