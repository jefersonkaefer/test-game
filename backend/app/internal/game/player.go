package game

import (
	"errors"

	"github.com/google/uuid"
)

type Player struct {
	id      uuid.UUID
	balance float64
	inPlay  bool
}

func NewPlayer() *Player {
	return &Player{
		id: uuid.New(),
	}
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

func (p *Player) StartGame(amount float64) error {
	if !p.CanBet(amount) {
		return errors.New("the player has no balance")
	}
	p.Debit(amount)
	p.inPlay = true
	return nil
}
