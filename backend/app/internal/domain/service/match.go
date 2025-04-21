package service

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"game/api/internal/application/repository"
	"game/api/internal/domain/entity"
	"game/api/internal/errs"
	"game/api/internal/infra/logger"
)

const (
	Even      = "even"
	Odd       = "odd"
	MaxNumber = 100
)

type MatchService struct {
	repoPlayer *repository.Players
	repoWallet *repository.Wallets
}

func NewMatchService(repoPlayer *repository.Players, repoWallet *repository.Wallets) *MatchService {
	return &MatchService{
		repoPlayer: repoPlayer,
		repoWallet: repoWallet,
	}
}

func (s *MatchService) NewMatch(ctx context.Context, clientID uuid.UUID) error {
	player, err := s.repoPlayer.Get(ctx, clientID)
	if err != nil {
		logger.Errorf("Failed to get player: %v", err)
		return err
	}
	if player.InPlay {
		return errs.ErrPlayerAlreadyInMatch
	}
	player.PlayOn()

	err = s.repoPlayer.Set(ctx, &player)
	if err != nil {
		logger.Errorf("Failed to set player in play: %v", err)
		return err
	}
	return nil
}

func (s *MatchService) PlaceBet(ctx context.Context, playerID uuid.UUID, amount float64, choice string) (number int, result string, err error) {
	player, err := s.repoPlayer.Get(ctx, playerID)
	if err != nil {
		logger.Errorf("Failed to get player: %v", err)
		return 0, "", err
	}
	if !player.HasBalance(amount) {
		return 0, "", errs.ErrInsufficientBalance
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	number = r.Intn(MaxNumber) + 1

	if number%2 == 0 {
		result = Even
	} else {
		result = Odd
	}

	if result == choice {
		player.Credit(amount)
		logger.Infof("Player %s won bet of %.2f", playerID, amount)
		result = "win"
	} else {
		player.Debit(amount)
		logger.Infof("Player %s lost bet of %.2f", playerID, amount)
		result = "lose"
	}

	err = s.repoPlayer.Set(ctx, &player)
	if err != nil {
		logger.Errorf("Failed to update player balance: %v", err)
		return 0, "", err
	}
	err = s.RefreshWallet(ctx, player)
	if err != nil {
		logger.Errorf("Failed to refresh wallet: %v", err)
		return 0, "", err
	}
	return number, result, nil
}

func (s *MatchService) RefreshWallet(ctx context.Context, player entity.Player) error {

	if !player.InPlay {
		return nil
	}

	w := entity.Wallet{
		ClientID: player.ClientID,
		Balance:  player.Balance,
	}
	err := s.repoWallet.Update(ctx, w)
	if err != nil {
		logger.Errorf("Failed to update wallet: %v", err)
		return err
	}
	return nil
}

func (s *MatchService) EndMatch(ctx context.Context, clientID uuid.UUID) error {
	player, err := s.repoPlayer.Get(ctx, clientID)
	if err != nil {
		logger.Errorf("Failed to get player: %v", err)
		return err
	}
	err = s.repoPlayer.EndGame(ctx, player.ClientID)
	if err != nil {
		logger.Errorf("Failed to clear player cache: %v", err)
		return err
	}
	logger.Infof("Match ended for player %s with final balance %.2f", clientID, player.Balance)
	return nil
}
