package controller

import (
	"context"
	"fmt"
	"time"

	"game/api/internal/application/dto"
	"game/api/internal/domain/entity"
	"game/api/internal/domain/service"
	"game/api/internal/infra/session"

	"github.com/google/uuid"
)

type MatchController struct {
	matchService *service.MatchService
}

func NewMatchController(matchService *service.MatchService) *MatchController {
	return &MatchController{
		matchService: matchService,
	}
}

func (c *MatchController) CreateMatch(ctx context.Context, clientID string, req dto.CreateMatchRequest) (res dto.CreateMatchResponse, err error) {
	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		return dto.CreateMatchResponse{}, fmt.Errorf("invalid client ID: %w", err)
	}

	match, err := c.matchService.CreateMatch(ctx,
		clientUUID,
		req.MinPlayers,
		req.MaxPlayers,
		entity.GameMode(req.GameMode),
	)
	if err != nil {
		return
	}

	playerIDs := make([]string, len(match.Players()))
	for i, p := range match.Players() {
		playerIDs[i] = p.PlayerID.String()
	}

	res = dto.CreateMatchResponse{
		ID:        match.ID().String(),
		Players:   playerIDs,
		Status:    string(match.Status()),
		CreatedAt: match.CreatedAt().Format(time.RFC3339),
	}
	return
}

func (c *MatchController) GetMatch(ctx context.Context, id string) (*dto.GetMatchResponse, error) {
	matchID, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	match, err := c.matchService.GetMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	playerIDs := make([]string, len(match.Players()))
	for i, p := range match.Players() {
		playerIDs[i] = p.PlayerID.String()
	}

	return &dto.GetMatchResponse{
		ID:        match.ID().String(),
		Players:   playerIDs,
		Status:    string(match.Status()),
		CreatedAt: match.CreatedAt().Format(time.RFC3339),
	}, nil
}

func (c *MatchController) JoinMatch(ctx context.Context, matchID, clientID string) (*dto.GetMatchResponse, error) {
	matchUUID, err := uuid.Parse(matchID)
	if err != nil {
		return nil, fmt.Errorf("invalid match ID: %w", err)
	}

	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		return nil, fmt.Errorf("invalid client ID: %w", err)
	}

	err = c.matchService.JoinMatch(ctx, matchUUID, clientUUID)
	if err != nil {
		return nil, err
	}

	match, err := c.matchService.GetMatch(ctx, matchUUID)
	if err != nil {
		return nil, err
	}

	playerIDs := make([]string, len(match.Players()))
	for i, p := range match.Players() {
		playerIDs[i] = p.PlayerID.String()
	}

	return &dto.GetMatchResponse{
		ID:        match.ID().String(),
		Players:   playerIDs,
		Status:    string(match.Status()),
		CreatedAt: match.CreatedAt().Format(time.RFC3339),
	}, nil
}

func (c *MatchController) LeaveMatch(ctx context.Context, req dto.AddPlayerRequest) error {
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		return fmt.Errorf("invalid match ID: %w", err)
	}

	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return fmt.Errorf("client ID is required")
	}

	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		return fmt.Errorf("invalid client ID: %w", err)
	}

	return c.matchService.LeaveMatch(ctx, matchID, clientUUID)
}

func (c *MatchController) PlaceBetAndChoose(ctx context.Context, clientID string, req dto.CreateBetRequest) error {
	matchID, err := uuid.Parse(req.MatchID)
	if err != nil {
		return fmt.Errorf("invalid match ID: %w", err)
	}

	playerID, err := uuid.Parse(clientID)
	if err != nil {
		return fmt.Errorf("invalid client ID: %w", err)
	}

	return c.matchService.PlaceBet(ctx, matchID, playerID, req.Amount, req.Parity)
}

func (c *MatchController) StartMatch(ctx context.Context, matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return fmt.Errorf("invalid match ID: %w", err)
	}
	return c.matchService.Play(ctx, id)
}

func (c *MatchController) EndMatch(ctx context.Context, matchID string) error {
	id, err := uuid.Parse(matchID)
	if err != nil {
		return fmt.Errorf("invalid match ID: %w", err)
	}
	return c.matchService.EndMatch(ctx, id)
}
