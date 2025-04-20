package entity

import (
	"game/api/internal/infra/database"

	"github.com/google/uuid"
)

type Wallet struct {
	guid    uuid.UUID
	balance float64
}

func NewWallet(balance float64) Wallet {
	return Wallet{
		guid:    uuid.New(),
		balance: balance,
	}
}

func (w *Wallet) GetID() uuid.UUID {
	return w.guid
}

func (w *Wallet) GetBalance() float64 {
	return w.balance
}

func (w *Wallet) Credit(amount float64) {
	w.balance += amount
}

func (w *Wallet) Debit(amount float64) {
	w.balance -= amount
}

func (w *Wallet) HasEnoughBalance(amount float64) bool {
	return w.balance >= amount
}

func LoadWallet(wData database.WalletData) (wallet Wallet, err error) {
	wallet.guid, err = uuid.Parse(wData.GUID)
	if err != nil {
		return
	}
	wallet.balance = wData.Balance

	return
}
