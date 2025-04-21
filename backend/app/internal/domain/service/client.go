package service

import (
	"context"

	"game/api/internal/application/repository"
	"game/api/internal/domain/entity"
	"game/api/internal/infra/logger"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ClientService struct {
	clientsRepo *repository.Clients
	walletRepo  *repository.Wallets
}

func NewClientService(clientsRepo *repository.Clients, walletRepo *repository.Wallets) *ClientService {
	return &ClientService{
		clientsRepo: clientsRepo,
		walletRepo:  walletRepo,
	}
}

func (s *ClientService) Create(ctx context.Context, username, password string) (clientID uuid.UUID, err error) {
	client, err := entity.NewClient(username, password, 0)
	if err != nil {
		return
	}

	err = s.clientsRepo.Add(client)
	if err != nil {
		return
	}
	wallet := entity.Wallet{
		ClientID: client.GetID(),
		Balance:  0,
	}
	err = s.walletRepo.Add(ctx, wallet)
	if err != nil {
		return
	}

	return client.GetID(), nil
}

func (s *ClientService) GetByUsername(ctx context.Context, username string) (client entity.Client, err error) {
	client, err = s.clientsRepo.GetByUsername(ctx, username)
	if err != nil {
		logger.Errorf("Failed to get client by username: %v", err)
		return
	}

	return client, nil
}

func (s *ClientService) GetBalance(ctx context.Context, clientID uuid.UUID) (balance float64, err error) {
	err = s.RefreshWallet(ctx, clientID)
	if err != nil {
		return
	}

	wallet, err := s.walletRepo.Get(ctx, clientID)
	if err != nil {
		return
	}

	return wallet.Balance, nil
}

func (s *ClientService) RefreshWallet(ctx context.Context, clientID uuid.UUID) error {
	err := s.walletRepo.ClearCache(ctx, clientID)
	if err != nil {
		logger.Errorf("Failed to clear wallet cache: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"client_id": clientID,
	}).Info("Wallet balance refreshed")

	return nil
}
