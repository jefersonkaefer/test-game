package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"game/api/internal/errs"
	"game/api/internal/infra/logger"

	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

type ClientData struct {
	GUID      string     `db:"guid"`
	Username  string     `db:"username"`
	Password  string     `db:"password"`
	CreatedAt string     `db:"created_at"`
	UpdatedAt string     `db:"updated_at"`
	DeletedAt *string    `db:"deleted_at"`
	Wallet    WalletData `db:"-"`
}

type WalletData struct {
	GUID      string  `db:"guid"`
	Balance   float64 `db:"balance"`
	ClientID  string  `db:"client_id"`
	CreatedAt string  `db:"created_at"`
	UpdatedAt string  `db:"updated_at"`
	DeletedAt *string `db:"deleted_at"`
}

type ClientWallet struct {
	ClientData
	WalletData
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

func (pg *Postgres) UpdateClient(client ClientData) error {
	logger.WithFields(logrus.Fields{
		"clientID": client.GUID,
	}).Debug("Updating client")

	query := fmt.Sprintf("UPDATE %s SET username = $1, password = $2, updated_at = $3 WHERE guid = $4", DB_TABLE_CLIENTS)
	_, err := pg.conn.Exec(
		query,
		client.Username,
		client.Password,
		client.UpdatedAt,
		client.GUID,
	)
	if err != nil {
		logger.Errorf("Failed to update client: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"clientID": client.GUID,
	}).Info("Client updated successfully")
	return nil
}

func (pg *Postgres) FindWalletByClientID(ctx context.Context, clientID string) (wallet WalletData, err error) {
	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Searching for wallet by client ID")

	err = pg.conn.QueryRowContext(ctx,
		fmt.Sprintf("SELECT guid, balance, client_id, created_at, updated_at FROM %s WHERE client_id = $1", DB_TABLE_WALLETS),
		clientID,
	).Scan(&wallet.GUID, &wallet.Balance, &wallet.ClientID, &wallet.CreatedAt, &wallet.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WithFields(logrus.Fields{
				"clientID": clientID,
			}).Warn("Wallet not found")
		}
	}

	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Wallet found successfully")
	return
}

func (pg *Postgres) UpdateWallet(ctx context.Context, wallet WalletData) error {
	logger.WithFields(logrus.Fields{
		"walletID": wallet.GUID,
	}).Debug("Updating wallet")

	query := fmt.Sprintf("UPDATE %s SET balance = $1, updated_at = $2 WHERE guid = $3 ", DB_TABLE_WALLETS)
	_, err := pg.conn.ExecContext(ctx,
		query,
		wallet.Balance,
		wallet.UpdatedAt,
		wallet.GUID,
	)
	if err != nil {
		logger.Errorf("Failed to update wallet: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"walletID": wallet.GUID,
	}).Info("Wallet updated successfully")
	return nil
}
