package database

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type postgres struct {
	conn *sql.DB
}

func NewPostgres(conn *sql.DB) *postgres {
	return &postgres{
		conn: conn,
	}
}
func (p *postgres) Insert() {

}
