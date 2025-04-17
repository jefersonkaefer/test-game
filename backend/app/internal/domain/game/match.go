package game

import (
	"errors"

	"game/api/internal/domain/entity"
)

type BetType string

const (
	Even BetType = "even"
	Odd  BetType = "odd"
)

type MatchResult struct {
	Number  int
	Bet     BetType
	Outcome bool
}
type MatchGame interface {
	DrawANumber() int
}
type match struct {
	game MatchGame
}

func NewMatch(game MatchGame) *match {
	return &match{
		game: game,
	}
}

func (m *match) Play(p entity.Client, amount float64, bet BetType) (mr MatchResult, err error) {
	defer p.PlayOff()
	if err != nil {
		return
	}
	if p.InPlay() {
		err = errors.New("player already in match")
		return
	}
	if !p.CanBet(amount) {
		err = errors.New("player have no balance")
		return
	}

	p.PlayOn()

	p.Debit(amount)

	mr.Bet = bet
	mr.Number = m.game.DrawANumber()
	mr.Outcome = CheckWon(mr.Number, bet)

	if mr.Outcome {
		p.Credit(amount)
	}
	return
}

func CheckWon(number int, bet BetType) bool {
	return (number%2 == 0 && bet == Even) || (number%2 != 0 && bet == Odd)
}
