package entity

import (
	"github.com/google/uuid"
)

type Player struct {
	id      uuid.UUID
	balance float64
	inPlay  bool
}

func NewPlayer() *Player {
	return &Player{
		id:      uuid.New(),
		balance: 1000,
		inPlay:  false,
	}
}

func (p *Player) GetID() uuid.UUID {
	return p.id
}

func (p *Player) CanBet(amount float64) bool {
	return p.balance >= amount
}

func (p *Player) InPlay() bool {
	return p.inPlay
}

func (p *Player) Debit(amount float64) {
	p.balance -= amount
}

func (p *Player) Credit(amount float64) {
	p.balance += amount
}

func (p *Player) PlayOn() {
	p.inPlay = true
}

func (p *Player) PlayOff() {
	p.inPlay = true
}
