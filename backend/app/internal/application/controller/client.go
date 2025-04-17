package controller

import (
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"game/api/internal/application/repository"
	"game/api/internal/domain/entity"
)

type Client struct {
	clientsRepo *repository.Client
}

func NewClient(clientsRepo *repository.Client) *Client {
	return &Client{
		clientsRepo: clientsRepo,
	}
}

type NewClientRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type NewClientResponse struct {
	ClientID string `json:"client_id"`
}

func (c *Client) NewClient(req NewClientRequest) (res NewClientResponse, err error) {
	client, err := entity.NewClient(req.Username, req.Password)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}
	err = c.clientsRepo.Add(client)
	if err != nil {
		log.Println("ERROR:", err)
	}
	res.ClientID = client.GetID().String()
	return
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

func (c *Client) Login(req LoginRequest) (res LoginResponse, err error) {
	client, err := c.clientsRepo.GetByUsername(req.Username)
	if err != nil {
		return
	}
	res.Token, err = GenerateJWT(client.GetID().String())
	if err != nil {
		log.Println("ERROR:", err)
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
