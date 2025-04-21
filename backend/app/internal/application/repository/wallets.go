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
	walletKeyPrefix = "wallet:"
)

type Wallets struct {
	cache *database.Redis
	db    *database.Postgres
}

func NewWallets(
	cache *database.Redis,
	db *database.Postgres,
) *Wallets {
	return &Wallets{
		cache: cache,
		db:    db,
	}
}

func (w *Wallets) getFromCacheOrDB(ctx context.Context, key string, fetchFromDB func() (database.WalletData, error)) (wData database.WalletData, err error) {
	lockKey := "lock:" + key
	err = w.cache.WithLock(ctx, lockKey, 5*time.Second, 3, 100*time.Millisecond, func() error {
		err = w.cache.Get(ctx, key, &wData)
		if err != nil && err == redis.Nil {
			wData, err = fetchFromDB()
			if err != nil {
				return err
			}

			err = w.cache.Set(ctx, key, wData)
			if err != nil {
				logger.Errorf("Failed to set wallet to cache: %v", err)
				return err
			}
		} else if err != nil {
			return err
		}
		return nil
	})
	return
}
func (w *Wallets) Add(ctx context.Context, wallet entity.Wallet) (err error) {
	key := walletKeyPrefix + wallet.ClientID.String()
	lockKey := "lock:" + key
	return w.cache.WithLock(ctx, lockKey, 5*time.Second, 3, 100*time.Millisecond, func() error {
		return w.db.InsertWallet(ctx, database.WalletData{
			GUID:     uuid.New().String(),
			Balance:  wallet.Balance,
			ClientID: wallet.ClientID.String(),
		})
	})
}
func (w *Wallets) Get(ctx context.Context, clientID uuid.UUID) (wallet entity.Wallet, err error) {
	key := walletKeyPrefix + clientID.String()
	var wData database.WalletData

	wData, err = w.getFromCacheOrDB(ctx, key, func() (database.WalletData, error) {
		return w.db.FindWalletByClientID(ctx, clientID.String())
	})
	if err != nil {
		return
	}

	wallet.ClientID, err = uuid.Parse(wData.ClientID)
	if err != nil {
		return
	}
	wallet.Balance = wData.Balance
	return
}

func (w *Wallets) Update(ctx context.Context, wallet entity.Wallet) (err error) {
	key := walletKeyPrefix + wallet.ClientID.String()
	lockKey := "lock:" + key
	walletData := database.WalletData{
		Balance:  wallet.Balance,
		ClientID: wallet.ClientID.String(),
	}

	return w.cache.WithLock(ctx, lockKey, 5*time.Second, 3, 100*time.Millisecond, func() error {
		err = w.db.UpdateWallet(ctx, walletData)
		if err != nil {
			logger.Errorf("Failed to update wallet in database: %v", err)
			return err
		}

		err = w.cache.Set(ctx, key, walletData)
		if err != nil {
			logger.Errorf("Failed to set wallet to cache: %v", err)
			return err
		}

		return nil
	})
}

func (w *Wallets) ClearCache(ctx context.Context, clientID uuid.UUID) (err error) {
	key := walletKeyPrefix + clientID.String()
	return w.cache.Delete(ctx, key)
}
