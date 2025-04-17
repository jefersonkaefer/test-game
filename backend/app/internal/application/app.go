package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	cache "github.com/redis/go-redis/v9"

	"game/api/internal/application/controller"
)

type App struct {
	ClientCtrl *controller.Client
}

func (a *App) NewGame() (uuid.UUID, error) {
	/*
			cfg := game.Config{
				MaxNumberDraw: 10,
			}
			g := game.NewGame(cfg)
			m := game.NewMatch(g)
			p := entity.NewClient()

			mr, err := m.Play(p, 10, game.Even)

			if err != nil {
				log.Println("ERRO:", err.Error())
			}
		fmt.Printf("%v", mr)
	*/
	return uuid.New(), nil
}

func DbConn() (*sql.DB, error) {
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USERNAME")
	pw := os.Getenv("POSTGRES_PASSWORD")
	db := os.Getenv("POSTGRES_DB")
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pw, db)
	return sql.Open("postgres", connStr)
}

func CacheConn() *cache.Client {
	addr := os.Getenv("REDIS_ADDR")
	password := "t3st"
	return cache.NewClient(&cache.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
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
	return res
}
