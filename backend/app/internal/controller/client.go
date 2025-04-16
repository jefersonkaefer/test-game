package controller

import (
	"game/api/internal/game/entity"
)

type NewClientRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) NewClient(req NewClientRequest) (err error, client entity.Client) {
	client = entity.NewClient()
	if err = a.db.InsertClient(client); err != nil {
		return
	}
	return
}
