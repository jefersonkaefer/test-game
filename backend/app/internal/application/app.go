package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	cache "github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"game/api/internal/application/controller"
	"game/api/internal/infra/database"
	"game/api/internal/infra/logger"
)

type App struct {
	ClientCtrl *controller.Client
}

func NewApp(ClientCtrl *controller.Client) *App {
	logger.Info("Initializing application")
	return &App{
		ClientCtrl: ClientCtrl,
	}
}

func (a *App) NewGame() (uuid.UUID, error) {
	logger.Debug("Creating new game")

	gameID := uuid.New()

	logger.WithFields(logrus.Fields{
		"gameID": gameID,
	}).Info("New game created successfully")

	return gameID, nil
}

func DbConn() (*database.Postgres, error) {
	logger.Debug("Establishing database connection")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USERNAME")
	pw := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DB")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pw, db)
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

func CacheConn() *database.Redis {
	logger.Debug("Establishing cache connection")
	addr := os.Getenv("REDIS_ADDR")
	password := "t3st"
	logger.Info("Cache connection established successfully")
	client := cache.NewClient(&cache.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	if err := client.Ping(context.Background()); err != nil {
		logger.Errorf("Failed to ping cache: %v", err)
		return nil
	}
	return database.NewRedis(client)
}

const (
	ActionNewPlayer = "new_player"
	ActionNewGame   = "new_game"
	ActionNewMatch  = "new_match"
)

type Request struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

type Response struct {
	Error string `json:"error"`
}

func (a *App) WebSocket(req Request) (res Response) {
	logger.WithFields(logrus.Fields{
		"action": req.Action,
	}).Debug("Processing WebSocket request")

	switch req.Action {
	case ActionNewPlayer:
		var r controller.NewClientRequest
		body, ok := req.Body.(map[string]interface{})
		if !ok {
			res.Error = "Invalid request body format"
			return res
		}
		jsonBody, _ := json.Marshal(body)
		if err := json.Unmarshal(jsonBody, &r); err != nil {
			log.Printf("Error parsing NewClientRequest: %v", err)
			res.Error = "Invalid request body"
			return res
		}
		_, err := a.ClientCtrl.NewClient(r)
		if err != nil {
			log.Printf("Error creating new client: %v", err)
			res.Error = err.Error()
			return res
		}
	default:
		res.Error = fmt.Sprintf("Unknown action: %s", req.Action)
	}

	logger.WithFields(logrus.Fields{
		"action": req.Action,
	}).Debug("WebSocket request processed successfully")

	return res
}
