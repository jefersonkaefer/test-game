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
	clientKeyPrefix = "client:"
)

type Clients struct {
	cache *database.Redis
	db    *database.Postgres
}

func NewClients(
	cache *database.Redis,
	db *database.Postgres,
) *Clients {
	return &Clients{
		cache: cache,
		db:    db,
	}
}

func (c *Clients) Add(client entity.Client) (err error) {
	cData := database.ClientData{
		GUID:     client.GetID().String(),
		Username: client.GetUsername(),
		Password: client.GetPassword(),
	}
	err = c.db.InsertClient(cData)
	if err != nil {
		logger.Errorf("Failed to insert client: %v", err)
	}
	return
}

func (c *Clients) getFromCacheOrDB(ctx context.Context, key string, fetchFromDB func() (database.ClientData, error)) (cData database.ClientData, err error) {
	lockKey := "lock:" + key
	err = c.cache.WithLock(ctx, lockKey, 5*time.Second, 3, 100*time.Millisecond, func() error {
		err = c.cache.Get(ctx, key, &cData)
		if err != nil && err == redis.Nil {
			cData, err = fetchFromDB()
			if err != nil {
				return err
			}

			err = c.cache.Set(ctx, key, cData)
			if err != nil {
				logger.Errorf("Failed to set client to cache: %v", err)
				return err
			}
		} else if err != nil {
			return err
		}
		return nil
	})
	return
}

func (c *Clients) GetByUsername(ctx context.Context, username string) (client entity.Client, err error) {
	key := clientKeyPrefix + username
	var cData database.ClientData

	cData, err = c.getFromCacheOrDB(ctx, key, func() (database.ClientData, error) {
		return c.db.FindClientByUsername(username)
	})
	if err != nil {
		return
	}

	client, err = entity.LoadClient(cData)
	return
}

func (c *Clients) Get(ctx context.Context, id uuid.UUID) (client entity.Client, err error) {
	key := clientKeyPrefix + id.String()
	var cData database.ClientData

	cData, err = c.getFromCacheOrDB(ctx, key, func() (database.ClientData, error) {
		return c.db.FindClientByID(id.String())
	})
	if err != nil {
		return
	}

	client, err = entity.LoadClient(cData)
	return
}

func (c *Clients) ClearCache(ctx context.Context, id uuid.UUID) (err error) {
	key := clientKeyPrefix + id.String()
	return c.cache.Delete(ctx, key)
}
