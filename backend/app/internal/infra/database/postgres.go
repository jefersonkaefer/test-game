package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"game/api/internal/errs"
	"game/api/internal/infra/logger"
)

const (
	DB_TABLE_CLIENTS = "clients"
	DB_TABLE_WALLETS = "wallets"
)

type Postgres struct {
	conn *sql.DB
}

func NewPostgres(conn *sql.DB) *Postgres {
	logger.Info("Initializing PostgreSQL connection")
	return &Postgres{
		conn: conn,
	}
}

func (pg *Postgres) InsertClient(c ClientData) error {
	logger.WithFields(logrus.Fields{
		"username": c.Username,
	}).Debug("Inserting new client")

	query := fmt.Sprintf("INSERT INTO %s (guid, username, password) VALUES ($1, $2, $3)", DB_TABLE_CLIENTS)
	_, err := pg.conn.Exec(
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

func (pg *Postgres) FindClientByUsername(username string) (client ClientData, err error) {
	logger.WithFields(logrus.Fields{
		"username": username,
	}).Debug("Searching for client by username")

	q := fmt.Sprintf(
		`SELECT c.guid, c.username, c.password, c.created_at, c.updated_at,
		w.guid, w.balance, w.client_id, w.created_at, w.updated_at 
		FROM %s c 
		INNER JOIN %s w ON c.guid = w.client_id and w.deleted_at IS NULL
		WHERE c.username = $1 and c.deleted_at IS NULL`,
		DB_TABLE_CLIENTS,
		DB_TABLE_WALLETS,
	)
	err = pg.conn.QueryRow(q, username).Scan(
		&client.GUID, &client.Username, &client.Password,
		&client.CreatedAt, &client.UpdatedAt,
		&client.Wallet.GUID, &client.Wallet.Balance, &client.Wallet.ClientID,
		&client.Wallet.CreatedAt, &client.Wallet.UpdatedAt,
	)
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

func (pg *Postgres) FindWalletByClientID(clientID string) (client ClientData, err error) {
	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Searching for client wallet")

	err = pg.conn.QueryRow(
		fmt.Sprintf("SELECT id, username, password FROM %s WHERE username = $1", DB_TABLE_WALLETS),
		clientID,
	).Scan(&client.GUID, &client.Username, &client.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WithFields(logrus.Fields{
				"clientID": clientID,
			}).Warn("Client wallet not found")
		} else {
			logger.Errorf("Failed to find client wallet: %v", err)
		}
		return
	}

	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Client wallet found successfully")
	return
}
func (pg *Postgres) Close() error {
	logger.Debug("Closing PostgreSQL connection")
	err := pg.conn.Close()
	if err != nil {
		logger.Errorf("Failed to close PostgreSQL connection: %v", err)
		return err
	}
	logger.Info("PostgreSQL connection closed successfully")
	return nil
}
