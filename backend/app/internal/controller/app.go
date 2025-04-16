package controller

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	"game/api/internal/game"
	"game/api/internal/game/entity"
	"game/api/internal/infra/database"
)

type App struct {
	db    *database.Postgres
	cache *database.Redis
}

func NewApp(
	db *database.Postgres,
	cache *database.Redis,
) *App {
	return &App{
		db:    db,
		cache: cache,
	}
}

func (a *App) NewGame() (uuid.UUID, error) {
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
	return uuid.New(), nil
}

type Request struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

type Response struct {
	Error string `json:"error"`
}
