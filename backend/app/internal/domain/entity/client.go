package entity

import (
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"game/api/internal/infra/database"
)

type Client struct {
	id       uuid.UUID
	username string
	password string
}

func NewClient(username, password string, balance float64) (Client, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		return Client{}, err
	}

	return Client{
		id:       uuid.New(),
		username: username,
		password: hashedPassword,
	}, nil
}

func (c *Client) GetID() uuid.UUID {
	return c.id
}

func (c *Client) GetUsername() string {
	return c.username
}

func (c *Client) GetPassword() string {
	return c.password
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (c *Client) CheckPasswordHash(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(c.password), []byte(password))
	return err == nil
}

func LoadClient(cData database.ClientData) (c Client, err error) {
	c.id, err = uuid.Parse(cData.GUID)
	if err != nil {
		return
	}
	c.username = cData.Username
	c.password = cData.Password
	return
}
