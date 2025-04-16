package database

import (
	cache "github.com/redis/go-redis/v9"
)

type redis struct {
	client *cache.Client
}

func NewRedis(client *cache.Client) *redis {
	return &redis{
		client: client,
	}
}
