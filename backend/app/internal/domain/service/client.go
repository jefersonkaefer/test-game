package service

import (
	"context"

	"game/api/internal/application/repository"
	"game/api/internal/domain/entity"
	"game/api/internal/infra/logger"

	"github.com/google/uuid"
)

type ClientService struct {
	clientsRepo *repository.Client
	walletRepo  *repository.Wallet
}

func NewClientService(clientsRepo *repository.Client, walletRepo *repository.Wallet) *ClientService {
	return &ClientService{
		clientsRepo: clientsRepo,
		walletRepo:  walletRepo,
	}
}

func (s *ClientService) Create(ctx context.Context, username, password string) (clientID string, err error) {
	client, err := entity.NewClient(username, password, 0)
	if err != nil {
		return
	}

	err = s.clientsRepo.Add(client)
	if err != nil {
		return
	}

	return client.GetID().String(), nil
}

func (s *ClientService) GetByUsername(ctx context.Context, username string) (client entity.Client, err error) {
	client, err = s.clientsRepo.GetByUsername(username)
	if err != nil {
		logger.Errorf("Failed to get client by username: %v", err)
		return
	}

	return client, nil
}

func (s *ClientService) Update(ctx context.Context, client entity.Client) error {
	return s.clientsRepo.Add(client)
}

func (s *ClientService) Credit(ctx context.Context, clientID uuid.UUID, amount float64) error {
	wallet, err := s.clientsRepo.GetWalletByClientID(ctx, clientID)
	if err != nil {
		return err
	}

	wallet.Credit(amount)

	return s.clientsRepo.UpdateWallet(ctx, wallet)
}

func (s *ClientService) Debit(ctx context.Context, clientID uuid.UUID, amount float64) error {
	wallet, err := s.clientsRepo.GetWalletByClientID(ctx, clientID)
	if err != nil {
		return err
	}

	wallet.Debit(amount)

	return s.clientsRepo.UpdateWallet(ctx, wallet)
}

func (s *ClientService) GetWallet(ctx context.Context, clientID uuid.UUID) (wallet entity.Wallet, err error) {
	wallet, err = s.walletRepo.GetByClientID(ctx, clientID)
	if err != nil {
		return
	}
	return
}
