package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"game/api/internal/errs"
	"game/api/internal/infra/logger"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type ClientData struct {
	GUID      string  `db:"guid" json:"guid"`
	Username  string  `db:"username" json:"username"`
	Password  string  `db:"password" json:"password"`
	CreatedAt string  `db:"created_at" json:"created_at"`
	UpdatedAt string  `db:"updated_at" json:"updated_at"`
	DeletedAt *string `db:"deleted_at" json:"deleted_at"`
}

func (c *ClientData) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}

func (c *ClientData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (pg *Postgres) InsertClient(c ClientData) error {
	logger.WithFields(logrus.Fields{
		"username": c.Username,
	}).Debug("Inserting new client")

	query := fmt.Sprintf("INSERT INTO %s (guid, username, password) VALUES ($1, $2, $3)", DB_TABLE_CLIENTS)
	_, err := pg.db.Exec(
		query,
		c.GUID,
		c.Username,
		c.Password,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			if pqErr.Code == "23505" {
				return errs.ErrUsernameExists
			}
		}
		logger.Errorf("Failed to insert client: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"username": c.Username,
	}).Info("Client inserted successfully")
	return nil
}

func (pg *Postgres) FindClientByID(clientID string) (client ClientData, err error) {
	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Searching for client by ID")

	q := fmt.Sprintf(
		`SELECT guid, username, password, created_at, updated_at, deleted_at
		FROM %s 
		WHERE guid = $1 and deleted_at IS NULL`,
		DB_TABLE_CLIENTS,
	)

	err = pg.db.Get(&client, q, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WithFields(logrus.Fields{
				"client_id": clientID,
			}).Warn("Client not found")
			err = errs.ErrNotFound
			return
		}
		logger.Errorf("Failed to find client: %v", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Client found successfully")
	return
}

func (pg *Postgres) FindClientByUsername(username string) (client ClientData, err error) {
	logger.WithFields(logrus.Fields{
		"username": username,
	}).Debug("Searching for client by username")

	q := fmt.Sprintf(
		`SELECT guid, username, password, created_at, updated_at, deleted_at
		FROM %s 
		WHERE username = $1 and deleted_at IS NULL`,
		DB_TABLE_CLIENTS,
	)

	err = pg.db.Get(&client, q, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WithFields(logrus.Fields{
				"username": username,
			}).Warn("Client not found")
			err = errs.ErrNotFound
			return
		}
		logger.Errorf("Failed to find client: %v", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"username": username,
	}).Debug("Client found successfully")
	return
}
