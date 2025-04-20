package game

import (
	"math/rand"
	"sync"
	"time"
)

type GameManager struct {
	players sync.Map
}

type Player struct {
	ID        string
	Balance   float64
	CurrentGame *Game
	mu        sync.Mutex
}

type Game struct {
	ID        string
	BetAmount float64
	BetType   string // "even" ou "odd"
	Result    int
	Won       bool
}

func NewGameManager() *GameManager {
	return &GameManager{}
}

func (gm *GameManager) GetPlayer(clientID string) *Player {
	player, _ := gm.players.LoadOrStore(clientID, &Player{
		ID:      clientID,
		Balance: 1000.0, // Saldo inicial
	})
	return player.(*Player)
}

func (p *Player) PlaceBet(amount float64, betType string) (*Game, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.CurrentGame != nil {
		return nil, ErrGameInProgress
	}

	if amount > p.Balance {
		return nil, ErrInsufficientBalance
	}

	// Gera número aleatório entre 1 e 6
	rand.Seed(time.Now().UnixNano())
	result := rand.Intn(6) + 1

	// Verifica se ganhou
	won := (result%2 == 0 && betType == "even") || (result%2 != 0 && betType == "odd")

	// Cria novo jogo
	game := &Game{
		ID:        time.Now().String(),
		BetAmount: amount,
		BetType:   betType,
		Result:    result,
		Won:       won,
	}

	p.CurrentGame = game
	p.Balance -= amount

	return game, nil
}

func (p *Player) EndGame() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.CurrentGame == nil {
		return ErrNoGameInProgress
	}

	if p.CurrentGame.Won {
		p.Balance += p.CurrentGame.BetAmount * 2
	}

	p.CurrentGame = nil
	return nil
}

// Erros personalizados
var (
	ErrGameInProgress     = errors.New("jogo em andamento")
	ErrNoGameInProgress   = errors.New("nenhum jogo em andamento")
	ErrInsufficientBalance = errors.New("saldo insuficiente")
) 