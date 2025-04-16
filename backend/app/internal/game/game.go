package game

import (
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"game/api/internal/game/entity"
)

type Config struct {
	MaxNumberDraw int //maximum limit of the number drawn.
}

type game struct {
	cfg     Config
	players map[string]*entity.Client
}

func NewGame(cfg Config) *game {
	return &game{
		cfg:     cfg,
		players: make(map[string]*entity.Client),
	}
}

func (g *game) DrawANumber() int {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	return rng.Intn(g.cfg.MaxNumberDraw) + 1
}

func (g *game) AddPlayer(p *entity.Client) {
	g.players[p.GetID().String()] = p
}

func (g *game) GetPlayer(playerUUID uuid.UUID) (*entity.Client, error) {
	uuid := playerUUID.String()
	if p, ok := g.players[uuid]; ok {
		return p, nil
	}

	return nil, errors.New("player not found")
}
