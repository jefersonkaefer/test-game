package entity

import "github.com/google/uuid"

type Player struct {
	ClientID uuid.UUID
	Balance  float64
	InPlay   bool
}

func (p *Player) PlayOn() {
	p.InPlay = true
}

func (p *Player) PlayOff() {
	p.InPlay = false
}

func (p *Player) GetBalance() float64 {
	return p.Balance
}

func (p *Player) Debit(amount float64) {
	p.Balance -= amount
}

func (p *Player) Credit(amount float64) {
	p.Balance += amount
}
func (p *Player) HasBalance(amount float64) bool {
	return p.Balance >= amount
}
