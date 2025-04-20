package service

import (
	"context"
	"fmt"

	"game/api/internal/application/repository"
	"game/api/internal/domain/entity"
	"game/api/internal/errs"
	"game/api/internal/infra/logger"

	"github.com/google/uuid"
)

type MatchService struct {
	matchRepo     *repository.Match
	clientService *ClientService
}

func NewMatchService(matchRepo *repository.Match, clientService *ClientService) *MatchService {
	return &MatchService{
		matchRepo:     matchRepo,
		clientService: clientService,
	}
}

func (s *MatchService) CreateMatch(
	ctx context.Context,
	playerID uuid.UUID,
	minPlayers,
	maxPlayers int,
	gameMode entity.GameMode,
) (m entity.Match, err error) {

	m = entity.NewMatch(
		playerID,
		minPlayers,
		maxPlayers,
		gameMode,
	)
	err = s.matchRepo.Create(ctx, m)
	if err != nil {
		logger.Errorf("failed to create match: %v", err)
		return
	}

	return
}

func (s *MatchService) GetMatch(ctx context.Context, matchID uuid.UUID) (m entity.Match, err error) {
	m, err = s.matchRepo.GetByID(ctx, matchID)
	return
}

func (s *MatchService) AddPlayer(ctx context.Context, matchID, playerID uuid.UUID) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	if err := match.AddPlayer(playerID); err != nil {
		return err
	}

	return s.matchRepo.Update(ctx, &match)
}

func (s *MatchService) RemovePlayer(ctx context.Context, matchID, playerID uuid.UUID) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	if err := match.RemovePlayer(playerID); err != nil {
		return err
	}

	return s.matchRepo.Update(ctx, &match)
}

func (s *MatchService) PlaceBet(ctx context.Context, matchID, playerID uuid.UUID, amount float64, choice string) error {
	// Verifica o saldo da carteira
	wallet, err := s.clientService.GetWallet(ctx, playerID)
	if err != nil {
		return err
	}

	if !wallet.HasEnoughBalance(amount) {
		return errs.ErrInsufficientBalance
	}

	// Busca a partida
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	// Adiciona a aposta e a escolha
	match.PlaceBet(playerID, amount, choice)

	return s.matchRepo.Update(ctx, &match)
}

func (s *MatchService) Play(ctx context.Context, matchID uuid.UUID) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	if err := match.Play(); err != nil {
		return err
	}

	return s.matchRepo.Update(ctx, &match)
}

func (s *MatchService) EndMatch(ctx context.Context, matchID uuid.UUID) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return err
	}

	match.End()

	return s.matchRepo.Update(ctx, &match)
}

func (s *MatchService) JoinMatch(ctx context.Context, matchID, playerID uuid.UUID) error {
	// Busca a partida
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	// Verifica se o jogador já está na partida
	for _, player := range match.Players() {
		if player.PlayerID == playerID {
			return errs.ErrPlayerAlreadyInMatch
		}
	}

	// Adiciona o jogador à partida
	if err := match.AddPlayer(playerID); err != nil {
		return fmt.Errorf("failed to add player: %w", err)
	}

	// Atualiza a partida no repositório
	if err := s.matchRepo.Update(ctx, &match); err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}

	return nil
}

func (s *MatchService) LeaveMatch(ctx context.Context, matchID, playerID uuid.UUID) error {
	// Busca a partida
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return fmt.Errorf("failed to get match: %w", err)
	}

	// Remove o jogador da partida
	if err := match.RemovePlayer(playerID); err != nil {
		return fmt.Errorf("failed to remove player: %w", err)
	}

	// Atualiza a partida no repositório
	if err := s.matchRepo.Update(ctx, &match); err != nil {
		return fmt.Errorf("failed to update match: %w", err)
	}

	return nil
}
