package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"game/api/internal/game/entity"
)

const (
	DB_TABLE_CLIENTS = "clients"
)

type Postgres struct {
	conn *sql.DB
}

func NewPostgres(conn *sql.DB) *Postgres {
	return &Postgres{
		conn: conn,
	}
}

func (pg *Postgres) InsertClient(p entity.Client) error {
	query := fmt.Sprintf("INSERT INTO %s (id, balance) VALUES ($1, $2)", DB_TABLE_CLIENTS)
	if _, err := pg.conn.Exec(
		query,
		p.GetID().String(),
		p.GetBalance(),
	); err != nil {
		return err
	}
	return nil
}
