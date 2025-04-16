package repository

import (
	"log"

	"game/api/internal/game/entity"
	"game/api/internal/infra/database"
)

type Client struct {
	db    *database.Postgres
	cache *database.Redis
}

func NewClient(
	db *database.Postgres,
	cache *database.Redis,
) *Client {
	return &Client{
		db:    db,
		cache: cache,
	}
}

func (c *Client) Add(client entity.Client) (err error) {
	cData := database.ClientData{
		GUID:     client.GetID().String(),
		Username: client.GetUsername(),
		Password: client.GetPassword(),
	}
	err = c.db.InsertClient(cData)
	if err != nil {
		log.Default().Println("ERROR:", err)
	}
	return
}

func (c *Client) GetByUsername(username string) (client entity.Client, err error) {
	cData, err := c.db.FindClientByUsername(username)
	if err != nil {
		return
	}
	client, err = entity.LoadClient(cData)
	return
}
