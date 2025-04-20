package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"game/api/internal/domain/entity"
	"game/api/internal/infra/database"
	"game/api/internal/infra/logger"
)

const (
	walletKeyPrefix = "wallet:"
)

type Wallet struct {
	db    *database.Postgres
	cache *redis.Client
}

func NewWallet(db *database.Postgres, cache *redis.Client) *Wallet {
	return &Wallet{
		db:    db,
		cache: cache,
	}
}

func (w *Wallet) GetByClientID(ctx context.Context, clientID uuid.UUID) (wallet entity.Wallet, err error) {
	key := walletKeyPrefix + clientID.String()
	var walletData database.WalletData
	walletDataJson, err := w.cache.Get(ctx, key).Bytes()
	if err != nil && err == redis.Nil {
		logger.Errorf("Failed to get wallet from cache: %v", err)
		walletData, err = w.db.FindWalletByClientID(ctx, clientID.String())
		if err != nil {
			return
		}
		w.cache.Set(ctx, key, walletData, 0)
	} else {
		if err := json.Unmarshal(walletDataJson, &walletData); err != nil {
			logger.Errorf("Failed to unmarshal wallet: %v", err)
		}
	}
	wallet, err = entity.LoadWallet(walletData)
	return
}

func (w *Wallet) Update(ctx context.Context, clientID uuid.UUID, wallet entity.Wallet) (err error) {
	walletData := database.WalletData{
		GUID:    wallet.GetID().String(),
		Balance: wallet.GetBalance(),
	}
	err = w.db.UpdateWallet(ctx, walletData)
	if err != nil {
		return
	}
	w.cache.Set(ctx, walletKeyPrefix+clientID.String(), walletData, 0)
	return
}
