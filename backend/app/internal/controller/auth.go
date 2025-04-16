package controller

import (
	"log"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (a *App) Login(req LoginRequest) (res LoginResponse, err error) {
	client, err := a.clients.GetByUsername(req.Username)
	if err != nil {
		return
	}
	log.Default().Println("client:", client)
	return
}
