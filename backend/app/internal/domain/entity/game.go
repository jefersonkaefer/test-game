package entity

import (
	"time"

	"github.com/google/uuid"
)

type GameMode string

const (
	GameModePlayer   GameMode = "player"   // Jogar contra um jogador específico
	GameModeAll      GameMode = "all"      // Jogar contra todos que quiserem
	GameModeComputer GameMode = "computer" // Jogar contra o computador
)

type MatchStatus string

const (
	MatchStatusWaiting  MatchStatus = "waiting"
	MatchStatusPlaying  MatchStatus = "playing"
	MatchStatusEnded    MatchStatus = "ended"
	MatchStatusFinished MatchStatus = "finished"
)

type PlayerRole string

const (
	PlayerRoleHost  PlayerRole = "host"
	PlayerRoleGuest PlayerRole = "guest"
)

const (
	Even           = "even"
	Odd            = "odd"
	PlayerComputer = "computer"
)

type Game struct {
	id          uuid.UUID
	creatorID   uuid.UUID
	name        string
	description string
	minPlayers  int
	maxPlayers  int
	createdAt   time.Time
	gameMode    GameMode
}

func (g *Game) ID() uuid.UUID {
	return g.id
}

func (g *Game) CreatorID() uuid.UUID {
	return g.creatorID
}

func (g *Game) Name() string {
	return g.name
}

func (g *Game) Description() string {
	return g.description
}

func (g *Game) MinPlayers() int {
	return g.minPlayers
}

func (g *Game) MaxPlayers() int {
	return g.maxPlayers
}

func (g *Game) CreatedAt() time.Time {
	return g.createdAt
}

func (g *Game) GameMode() GameMode {
	return g.gameMode
}

func (g *Game) SetGameMode(mode GameMode) {
	g.gameMode = mode
}

func NewGame(creatorID uuid.UUID, name, description string, minPlayers, maxPlayers int, gameMode GameMode) Game {
	return Game{
		id:          uuid.New(),
		creatorID:   creatorID,
		name:        name,
		description: description,
		minPlayers:  minPlayers,
		maxPlayers:  maxPlayers,
		createdAt:   time.Now(),
		gameMode:    gameMode,
	}
}
