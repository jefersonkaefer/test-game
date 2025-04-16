package database

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/lib/pq"
	_ "github.com/lib/pq"

	"game/api/internal/errs"
)

const (
	DB_TABLE_CLIENTS = "clients"
	DB_TABLE_WALLETS = "wallets"
)

type Postgres struct {
	conn *sql.DB
}

func NewPostgres(conn *sql.DB) *Postgres {
	return &Postgres{
		conn: conn,
	}
}
func (pg *Postgres) InsertClient(c ClientData) error {
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
		return err
	}
	return nil
}

func (pg *Postgres) FindClientByUsername(username string) (client ClientData, err error) {
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
		log.Default().Println("%v", err)
		if errors.Is(err, sql.ErrNoRows) {
			err = errs.ErrNotFound
			return
		}
		return
	}
	return
}

func (pg *Postgres) FindWalletByClientID(clientID string) (client ClientData, err error) {
	err = pg.conn.QueryRow(
		fmt.Sprintf("SELECT id, username, password FROM %s WHERE username = $1", DB_TABLE_WALLETS),
		clientID,
	).Scan(&client.GUID, &client.Username, &client.Password)
	return
}
