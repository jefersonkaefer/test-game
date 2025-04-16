package controller

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	res.Token, err = GenerateJWT(client.GetID().String())
	if err != nil {
		log.Default().Println("ERROR:", err)
	}
	return
}

func GenerateJWT(clientId string) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET_KEY"))

	claims := jwt.MapClaims{
		"client_id": clientId,
		"iss":       "game-api",
		"exp":       time.Now().Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
