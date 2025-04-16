package entity

import (
	"github.com/google/uuid"
)

type Wallet struct {
	guid    uuid.UUID
	balance float64
}

type Client struct {
	guid   uuid.UUID
	inPlay bool
	wallet Wallet
}

func NewClient() Client {
	return Client{
		guid:   uuid.New(),
		inPlay: false,
		wallet: Wallet{
			guid:    uuid.New(),
			balance: 0,
		},
	}
}

func (p *Client) GetID() uuid.UUID {
	return p.guid
}

func (p *Client) GetBalance() float64 {
	return p.wallet.balance
}

func (p *Client) CanBet(amount float64) bool {
	return p.wallet.balance >= amount
}

func (p *Client) InPlay() bool {
	return p.inPlay
}

func (p *Client) Debit(amount float64) {
	p.wallet.balance -= amount
}

func (p *Client) Credit(amount float64) {
	p.wallet.balance += amount
}

func (p *Client) PlayOn() {
	p.inPlay = true
}

func (p *Client) PlayOff() {
	p.inPlay = true
}
