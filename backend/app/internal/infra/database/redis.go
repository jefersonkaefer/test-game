package database

import (
	cache "github.com/redis/go-redis/v9"
)

type Redis struct {
	client *cache.Client
}

func NewRedis(client *cache.Client) *Redis {
	return &Redis{
		client: client,
	}
}
