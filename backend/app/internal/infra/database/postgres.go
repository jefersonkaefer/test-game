package database

import (
	"database/sql"

	_ "github.com/lib/pq"

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
