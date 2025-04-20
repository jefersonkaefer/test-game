package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"game/api/internal/domain/entity"
	"game/api/internal/infra/database"
	"game/api/internal/infra/logger"
)

type Client struct {
	db    *database.Postgres
	cache *redis.Client
}

func NewClient(
	db *database.Postgres,
	cache *redis.Client,
) *Client {
	return &Client{
		db:    db,
		cache: cache,
	}
}

func (c *Client) Add(client entity.Client) (err error) {
	cData := database.ClientData{
		GUID:     client.GetID().String(),
		Username: client.GetUsername(),
		Password: client.GetPassword(),
		Wallet: database.WalletData{
			GUID:     client.GetWalletID().String(),
			Balance:  client.GetBalance(),
			ClientID: client.GetID().String(),
		},
	}
	err = c.db.InsertClient(cData)
	if err != nil {
		logger.Errorf("Failed to insert client: %v", err)
	}
	return
}

func (c *Client) GetByUsername(username string) (client entity.Client, err error) {
	cData, err := c.db.FindClientByUsername(username)
	if err != nil {
		logger.Errorf("Failed to find client by username: %v", err)
		return
	}
	client, err = entity.LoadClient(cData)
	return
}

func (c *Client) Update(client entity.Client) (err error) {
	cData := database.ClientData{
		GUID:     client.GetID().String(),
		Username: client.GetUsername(),
		Password: client.GetPassword(),
	}
	err = c.db.UpdateClient(cData)
	if err != nil {
		logger.Errorf("Failed to update client: %v", err)
	}
	return
}

func (c *Client) GetWalletByClientID(ctx context.Context, clientID uuid.UUID) (wallet entity.Wallet, err error) {
	wData, err := c.db.FindWalletByClientID(ctx, clientID.String())
	if err != nil {
		logger.Errorf("Failed to find wallet by client ID: %v", err)
		return
	}
	wallet, err = entity.LoadWallet(wData)
	return
}

func (c *Client) UpdateWallet(ctx context.Context, wallet entity.Wallet) (err error) {
	wData := database.WalletData{
		GUID:    wallet.GetID().String(),
		Balance: wallet.GetBalance(),
	}
	err = c.db.UpdateWallet(ctx, wData)
	if err != nil {
		logger.Errorf("Failed to update wallet: %v", err)
	}
	return
}
