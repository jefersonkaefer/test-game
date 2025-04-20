package application

import (
	"context"
	"database/sql"
	"fmt"
	"game/api/internal/infra/database"
	"game/api/internal/infra/logger"
	"os"

	cache "github.com/redis/go-redis/v9"
)

const (
	ActionNewPlayer    = "new_player"
	ActionCreateGame   = "create_game" // Criar novo jogo
	ActionJoinMatch    = "join_match"
	ActionLeaveMatch   = "leave_match"
	ActionChooseParity = "choose_parity" // Escolher se é ímpar ou par
	ActionPlaceBet     = "place_bet"     // Fazer uma aposta
	ActionMatchResult  = "match_result"  // Resultado da partida
)

type Request struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

type Response struct {
	Error string      `json:"error"`
	Data  interface{} `json:"data"`
}

func DbConn(ctx context.Context) (*database.Postgres, error) {
	logger.Debug("Establishing database connection")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USERNAME"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)
	logger.Info("Database connection established successfully")
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		logger.Errorf("Failed to connect to database: %v", err)
		return nil, err
	}
	if err := conn.Ping(); err != nil {
		logger.Errorf("Failed to ping database: %v", err)
		return nil, err
	}
	logger.Info("Database connection pinged successfully")
	return database.NewPostgres(conn), nil
}

func RedisConn(ctx context.Context) *cache.Client {
	logger.Debug("Establishing cache connection")
	logger.Info("Cache connection established successfully")
	client := cache.NewClient(&cache.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		logger.Errorf("Failed to ping cache: %v", err)
		return nil
	}
	return client
}
