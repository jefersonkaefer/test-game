package database

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
