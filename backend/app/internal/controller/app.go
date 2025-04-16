package controller

import (
	"github.com/google/uuid"

	"game/api/internal/repository"
)

type App struct {
	clients *repository.Client
}

func NewApp(clients *repository.Client) *App {
	return &App{
		clients: clients,
	}
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

type Request struct {
	Action string      `json:"action"`
	Body   interface{} `json:"body"`
}

type Response struct {
	Error string `json:"error"`
}
