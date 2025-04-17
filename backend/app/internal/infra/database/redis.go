package database

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"game/api/internal/infra/logger"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) *Redis {
	logger.Info("Initializing Redis connection")
	return &Redis{client: client}
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}) error {
	logger.WithFields(logrus.Fields{
		"key": key,
	}).Debug("Setting Redis key")

	data, err := json.Marshal(value)
	if err != nil {
		logger.Errorf("Failed to marshal value: %v", err)
		return err
	}

	err = r.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		logger.Errorf("Failed to set Redis key: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"key": key,
	}).Debug("Redis key set successfully")
	return nil
}

func (r *Redis) Get(ctx context.Context, key string, value interface{}) error {
	logger.WithFields(logrus.Fields{
		"key": key,
	}).Debug("Getting Redis key")

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.WithFields(logrus.Fields{
				"key": key,
			}).Warn("Redis key not found")
		} else {
			logger.Errorf("Failed to get Redis key: %v", err)
		}
		return err
	}

	err = json.Unmarshal(data, value)
	if err != nil {
		logger.Errorf("Failed to unmarshal value: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"key": key,
	}).Debug("Redis key retrieved successfully")
	return nil
}

func (r *Redis) Delete(ctx context.Context, key string) error {
	logger.WithFields(logrus.Fields{
		"key": key,
	}).Debug("Deleting Redis key")

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		logger.Errorf("Failed to delete Redis key: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"key": key,
	}).Debug("Redis key deleted successfully")
	return nil
}

func (r *Redis) Close() error {
	logger.Debug("Closing Redis connection")
	err := r.client.Close()
	if err != nil {
		logger.Errorf("Failed to close Redis connection: %v", err)
		return err
	}
	logger.Debug("Redis connection closed successfully")
	return nil
}
