package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"game/api/internal/errs"
	"game/api/internal/infra/logger"

	"github.com/sirupsen/logrus"
)

type WalletData struct {
	GUID      string  `db:"guid" json:"guid"`
	ClientID  string  `db:"client_id" json:"client_id"`
	Balance   float64 `db:"balance" json:"balance"`
	CreatedAt string  `db:"created_at" json:"created_at"`
	UpdatedAt string  `db:"updated_at" json:"updated_at"`
	DeletedAt *string `db:"deleted_at" json:"deleted_at"`
}

func (w *WalletData) MarshalBinary() ([]byte, error) {
	return json.Marshal(w)
}

func (w *WalletData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, w)
}

func (pg *Postgres) FindWalletByClientID(ctx context.Context, clientID string) (wallet WalletData, err error) {
	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Searching for wallet by client ID")

	q := fmt.Sprintf(
		`SELECT guid, balance, client_id, created_at, updated_at 
		FROM %s 
		WHERE client_id = $1 and deleted_at IS NULL`,
		DB_TABLE_WALLETS,
	)

	err = pg.db.GetContext(ctx, &wallet, q, clientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			logger.WithFields(logrus.Fields{
				"clientID": clientID,
			}).Warn("Wallet not found")
			err = errs.ErrNotFound
			return
		}
		logger.Errorf("Failed to find wallet: %v", err)
		return
	}

	logger.WithFields(logrus.Fields{
		"clientID": clientID,
	}).Debug("Wallet found successfully")
	return
}
func (pg *Postgres) InsertWallet(ctx context.Context, wallet WalletData) error {
	logger.WithFields(logrus.Fields{
		"walletID": wallet.GUID,
	}).Debug("Inserting wallet")

	query := fmt.Sprintf(
		"INSERT INTO %s (guid, balance, client_id, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())",
		DB_TABLE_WALLETS,
	)

	_, err := pg.db.ExecContext(ctx, query, wallet.GUID, wallet.Balance, wallet.ClientID)
	if err != nil {
		logger.Errorf("Failed to insert wallet: %v", err)
		return err
	}

	logger.WithFields(logrus.Fields{
		"walletID": wallet.GUID,
	}).Info("Wallet inserted successfully")
	return nil
}

func (pg *Postgres) UpdateWallet(ctx context.Context, wallet WalletData) error {
	logger.WithFields(logrus.Fields{
		"walletID": wallet.GUID,
	}).Debug("Updating wallet")

	query := fmt.Sprintf(
		"UPDATE %s SET balance = $1, updated_at = NOW() WHERE client_id = $2",
		DB_TABLE_WALLETS,
	)

	_, err := pg.db.ExecContext(ctx,
		query,
		wallet.Balance,
		wallet.ClientID,
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
