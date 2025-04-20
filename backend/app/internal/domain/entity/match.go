package entity

import (
	"math/rand"
	"time"

	"github.com/google/uuid"

	"game/api/internal/errs"
)

type MatchPlayer struct {
	PlayerID uuid.UUID
	Role     PlayerRole
}

type Match struct {
	id          uuid.UUID
	creatorID   uuid.UUID
	minPlayers  int
	maxPlayers  int
	gameMode    GameMode
	players     []MatchPlayer
	status      MatchStatus
	createdAt   time.Time
	choices     map[string]string
	bets        map[string]float64
	result      string
	currentTurn string
}

func NewMatch(
	creatorID uuid.UUID,
	minPlayers int,
	maxPlayers int,
	gameMode GameMode,
) Match {
	return Match{
		id:         uuid.New(),
		creatorID:  creatorID,
		minPlayers: minPlayers,
		maxPlayers: maxPlayers,
		gameMode:   gameMode,
		status:     MatchStatusWaiting,
		createdAt:  time.Now(),
		players:    make([]MatchPlayer, 0),
		choices:    make(map[string]string),
		bets:       make(map[string]float64),
	}
}

func (m *Match) ID() uuid.UUID {
	return m.id
}

func (m *Match) CreatorID() uuid.UUID {
	return m.creatorID
}

func (m *Match) MinPlayers() int {
	return m.minPlayers
}

func (m *Match) MaxPlayers() int {
	return m.maxPlayers
}

func (m *Match) GameMode() GameMode {
	return m.gameMode
}

func (m *Match) Players() []MatchPlayer {
	return m.players
}

func (m *Match) Status() MatchStatus {
	return m.status
}

func (m *Match) SetStatus(status MatchStatus) {
	m.status = status
}

func (m *Match) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Match) Choices() map[string]string {
	return m.choices
}

func (m *Match) SetChoices(choices map[string]string) {
	m.choices = choices
}

func (m *Match) Bets() map[string]float64 {
	return m.bets
}

func (m *Match) SetBets(bets map[string]float64) {
	m.bets = bets
}

func (m *Match) Result() string {
	return m.result
}

func (m *Match) CurrentTurn() string {
	return m.currentTurn
}

func (m *Match) SetCurrentTurn(currentTurn string) {
	m.currentTurn = currentTurn
}

func (m *Match) AddPlayer(playerID uuid.UUID) error {
	for _, player := range m.players {
		if player.PlayerID == playerID {
			return errs.ErrPlayerAlreadyInMatch
		}
	}

	if len(m.players) >= m.maxPlayers {
		return errs.ErrMatchFull
	}

	if m.status != MatchStatusWaiting {
		return errs.ErrMatchNotJoinable
	}

	role := PlayerRoleGuest
	if len(m.players) == 0 {
		role = PlayerRoleHost
	}

	m.players = append(m.players, MatchPlayer{
		PlayerID: playerID,
		Role:     role,
	})
	return nil
}

func (m *Match) RemovePlayer(playerID uuid.UUID) error {
	if m.status != MatchStatusWaiting {
		return errs.ErrMatchNotLeavable
	}

	for i, player := range m.players {
		if player.PlayerID.String() == playerID.String() {
			m.players = append(m.players[:i], m.players[i+1:]...)
			break
		}
	}

	if len(m.players) == 0 {
		return nil
	}

	if m.players[0].Role != PlayerRoleHost {
		m.players[0] = MatchPlayer{
			PlayerID: m.players[0].PlayerID,
			Role:     PlayerRoleHost,
		}
	}

	return nil
}

func (m *Match) Play() error {
	lenPlayers := len(m.players)
	if lenPlayers < m.minPlayers {
		return errs.ErrMatchMinPlayers
	}

	if lenPlayers > m.maxPlayers {
		return errs.ErrMatchMaxPlayers
	}
	m.status = MatchStatusPlaying
	return nil
}

func (m *Match) End() {
	m.status = MatchStatusEnded
}

func (m *Match) PlaceBet(playerID uuid.UUID, amount float64, choice string) {
	// Inicializa os mapas se necessário
	if m.bets == nil {
		m.bets = make(map[string]float64)
	}
	if m.choices == nil {
		m.choices = make(map[string]string)
	}

	// Adiciona ou atualiza a aposta
	if existingAmount, ok := m.bets[playerID.String()]; ok {
		m.bets[playerID.String()] = existingAmount + amount
	} else {
		m.bets[playerID.String()] = amount
	}

	// Adiciona a escolha
	m.choices[playerID.String()] = choice

	// Se for modo computador e for a primeira escolha, faz a escolha do computador
	if m.gameMode == GameModeComputer && len(m.choices) == 1 {
		aiChoice := Even
		if choice == Even {
			aiChoice = Odd
		}
		m.choices[PlayerComputer] = aiChoice
		m.status = MatchStatusFinished
	}
}

func (m *Match) CalculateWinners() map[string]float64 {
	winners := make(map[string]float64)

	for playerID, bet := range m.bets {
		if m.choices[playerID] == m.result {
			winners[playerID] = bet
		}
	}

	return winners
}

func (m *Match) BothPlayersChose() bool {
	if m.gameMode == GameModeComputer {
		return len(m.choices) == 2
	}
	if m.gameMode == GameModePlayer {
		if _, ok := m.choices[m.currentTurn]; !ok {
			return false
		}
		return true
	}
	for _, p := range m.players {
		if _, ok := m.choices[p.PlayerID.String()]; !ok {
			return false
		}
	}
	return true
}

func (m *Match) Run() string {
	// Gera um número aleatório entre 0 e 100
	numero := rand.Intn(101) // 101 porque Intn gera um número entre 0 e n-1

	// Se o número for par, retorna Even
	if numero%2 == 0 {
		return Even
	}

	// Se o número for ímpar, retorna Odd
	return Odd
}

func LoadMatch(
	id string,
	minPlayers int,
	maxPlayers int,
	gameMode GameMode,
	players []MatchPlayer,
	status MatchStatus,
	createdAt time.Time,
	choices map[string]string,
	bets map[string]float64,
	result string,
	currentTurn string,
) (m Match, err error) {
	matchID, err := uuid.Parse(id)
	if err != nil {
		return Match{}, err
	}

	m = Match{
		id:          matchID,
		minPlayers:  minPlayers,
		maxPlayers:  maxPlayers,
		gameMode:    gameMode,
		players:     players,
		status:      status,
		createdAt:   createdAt,
		choices:     choices,
		bets:        bets,
		result:      result,
		currentTurn: currentTurn,
	}
	return m, nil
}

func (m *Match) SetPlayers(players []MatchPlayer) {
	m.players = players
}
