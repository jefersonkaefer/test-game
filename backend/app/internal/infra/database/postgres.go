package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"game/api/internal/infra/logger"
)

const (
	DB_TABLE_CLIENTS = "clients"
	DB_TABLE_WALLETS = "wallets"
)

type Postgres struct {
	db *sqlx.DB
}

func NewPostgres(db *sqlx.DB) *Postgres {
	logger.Info("Initializing PostgreSQL connection")
	return &Postgres{
		db: db,
	}
}

func (pg *Postgres) Close() error {
	logger.Debug("Closing PostgreSQL connection")
	err := pg.db.Close()
	if err != nil {
		logger.Errorf("Failed to close PostgreSQL connection: %v", err)
		return err
	}
	logger.Info("PostgreSQL connection closed successfully")
	return nil
}
