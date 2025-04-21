package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"game/api/internal/domain/entity"
	"game/api/internal/infra/database"
	"game/api/internal/infra/logger"
)

const (
	playerKeyPrefix = "player:"
)

type Players struct {
	cache      *database.Redis
	repoClient *Clients
	repoWallet *Wallets
}

func NewPlayers(
	cache *database.Redis,
	repoClient *Clients,
	repoWallet *Wallets,
) *Players {
	return &Players{
		cache:      cache,
		repoClient: repoClient,
		repoWallet: repoWallet,
	}
}

func (p *Players) getFromCacheOrDB(ctx context.Context, key string, fetchFromDB func() (database.PlayerData, error)) (pData database.PlayerData, err error) {
	lockKey := "lock:" + key
	err = p.cache.WithLock(ctx, lockKey, 5*time.Second, 3, 100*time.Millisecond, func() error {
		err = p.cache.Get(ctx, key, &pData)
		if err != nil && err == redis.Nil {
			pData, err = fetchFromDB()
			if err != nil {
				return err
			}

			err = p.cache.Set(ctx, key, pData)
			if err != nil {
				logger.Errorf("Failed to set player to cache: %v", err)
				return err
			}
		} else if err != nil {
			return err
		}
		return nil
	})
	return
}

func (p *Players) Get(ctx context.Context, clientID uuid.UUID) (entity.Player, error) {
	key := playerKeyPrefix + clientID.String()
	var pData database.PlayerData

	pData, err := p.getFromCacheOrDB(ctx, key, func() (database.PlayerData, error) {
		client, err := p.repoClient.Get(ctx, clientID)
		if err != nil {
			logger.Errorf("Error getting client from repository: %v", err)
			return database.PlayerData{}, err
		}

		wallet, err := p.repoWallet.Get(ctx, clientID)
		if err != nil {
			logger.Errorf("Error getting wallet from repository: %v", err)
			return database.PlayerData{}, err
		}

		return database.PlayerData{
			ClientID: client.GetID().String(),
			Balance:  wallet.Balance,
			InPlay:   false,
		}, nil
	})
	if err != nil {
		return entity.Player{}, err
	}

	clientUUID, err := uuid.Parse(pData.ClientID)
	if err != nil {
		logger.Errorf("Failed to parse client ID: %v", err)
		return entity.Player{}, err
	}

	return entity.Player{
		ClientID: clientUUID,
		Balance:  pData.Balance,
		InPlay:   pData.InPlay,
	}, nil
}

func (p *Players) Set(ctx context.Context, player *entity.Player) error {
	key := playerKeyPrefix + player.ClientID.String()
	lockKey := "lock:" + key
	playerData := database.PlayerData{
		ClientID: player.ClientID.String(),
		Balance:  player.Balance,
		InPlay:   player.InPlay,
	}

	return p.cache.WithLock(ctx, lockKey, 5*time.Second, 3, 100*time.Millisecond, func() error {
		err := p.cache.Set(ctx, key, playerData)
		if err != nil {
			logger.Errorf("Failed to set player to cache: %v", err)
			return err
		}
		return nil
	})
}

func (p *Players) EndGame(ctx context.Context, playerID uuid.UUID) error {
	key := playerKeyPrefix + playerID.String()
	lockKey := "lock:" + key
	return p.cache.WithLock(ctx, lockKey, 5*time.Second, 3, 100*time.Millisecond, func() error {
		err := p.cache.Delete(ctx, key)
		if err != nil {
			return err
		}
		return nil
	})
}

func (p *Players) ClearCache(ctx context.Context, clientID uuid.UUID) (err error) {
	key := playerKeyPrefix + clientID.String()
	return p.cache.Delete(ctx, key)
}
