package lock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"game/api/internal/infra/logger"
)

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockExpired     = errors.New("lock expired")
)

type Lock struct {
	client *redis.Client
}

func NewLock(client *redis.Client) *Lock {
	return &Lock{client: client}
}

// tenta adquirir um lock  com timeout e retry
func (l *Lock) Acquire(ctx context.Context, key string, ttl time.Duration, retryCount int, retryDelay time.Duration) (bool, error) {
	for i := 0; i < retryCount; i++ {
		// Tenta adquirir o lock
		acquired, err := l.client.SetNX(ctx, key, "1", ttl).Result()
		if err != nil {
			return false, err
		}

		if acquired {
			logger.WithFields(logrus.Fields{
				"key": key,
			}).Debug("Lock acquired successfully")
			return true, nil
		}

		// verifica se o lock ainda existe
		exists, err := l.client.Exists(ctx, key).Result()
		if err != nil {
			return false, err
		}

		if exists == 0 {
			// expirado então tenta novamente
			continue
		}

		// espera antes de tentar novamente
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-time.After(retryDelay):
			continue
		}
	}

	logger.WithFields(logrus.Fields{
		"key": key,
	}).Warn("Failed to acquire lock after retries")
	return false, ErrLockNotAcquired
}

// libera o lock
func (l *Lock) Release(ctx context.Context, key string) error {
	// garante que apenas o dono libere
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	_, err := l.client.Eval(ctx, script, []string{key}, "1").Result()
	if err != nil {
		logger.WithFields(logrus.Fields{
			"key": key,
			"err": err,
		}).Error("Failed to release lock")
		return err
	}

	logger.WithFields(logrus.Fields{
		"key": key,
	}).Debug("Lock released successfully")
	return nil
}

// executa a função com um lock
func (l *Lock) WithLock(ctx context.Context, key string, ttl time.Duration, retryCount int, retryDelay time.Duration, fn func() error) error {
	acquired, err := l.Acquire(ctx, key, ttl, retryCount, retryDelay)
	if err != nil {
		return err
	}

	if !acquired {
		return ErrLockNotAcquired
	}

	defer func() {
		_ = l.Release(ctx, key)
	}()

	return fn()
}

type Locker struct {
	redis      *redis.Client
	expiration time.Duration
	retryDelay time.Duration
	maxRetries int
}

func NewLocker(redis *redis.Client, expiration time.Duration, retryDelay time.Duration, maxRetries int) *Locker {
	return &Locker{
		redis:      redis,
		expiration: expiration,
		retryDelay: retryDelay,
		maxRetries: maxRetries,
	}
}

func (l *Locker) Lock(ctx context.Context, key, value string) error {
	for i := 0; i < l.maxRetries; i++ {
		success, err := l.redis.SetNX(ctx, key, value, l.expiration).Result()
		if err != nil {
			return fmt.Errorf("failed to acquire lock: %w", err)
		}

		if success {
			return nil
		}

		exists, err := l.redis.Exists(ctx, key).Result()
		if err != nil {
			return fmt.Errorf("failed to check if lock exists: %w", err)
		}

		if exists == 0 {
			continue
		}

		logger.Debugf("Lock for key %s already exists, retrying in %s", key, l.retryDelay)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(l.retryDelay):
			continue
		}
	}

	return fmt.Errorf("failed to acquire lock after %d retries", l.maxRetries)
}

func (l *Locker) Unlock(ctx context.Context, key, value string) error {
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.redis.Eval(ctx, script, []string{key}, value).Result()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}

	if result.(int64) != 1 {
		return fmt.Errorf("failed to release lock: lock does not exist or was acquired by another client")
	}

	return nil
}

func (l *Locker) WithLock(ctx context.Context, key, value string, fn func() error) error {
	err := l.Lock(ctx, key, value)
	if err != nil {
		return err
	}

	defer func() {
		if err := l.Unlock(ctx, key, value); err != nil {
			logger.Errorf("Failed to release lock: %v", err)
		}
	}()

	return fn()
}
