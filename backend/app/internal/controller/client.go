package controller

import (
	"log"

	"game/api/internal/game/entity"
)

type NewClientRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NewClientResponse struct {
	ClientID string `json:"client_id"`
}

func (a *App) NewClient(req NewClientRequest) (res NewClientResponse, err error) {
	client, err := entity.NewClient(req.Username, req.Password)
	if err != nil {
		log.Default().Println("ERROR:", err)
		return
	}
	err = a.clients.Add(client)
	if err != nil {
		log.Default().Println("ERROR:", err)
	}
	res.ClientID = client.GetID().String()
	return
}
