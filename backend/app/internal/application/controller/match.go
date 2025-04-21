package controller

import (
	"context"

	"github.com/google/uuid"

	"game/api/internal/application/dto"
	"game/api/internal/domain/service"
	"game/api/internal/infra/logger"
)

type MatchController struct {
	serviceMatch *service.MatchService
}

func NewMatchController(serviceMatch *service.MatchService) *MatchController {
	return &MatchController{
		serviceMatch: serviceMatch,
	}
}

func (c *MatchController) NewMatch(ctx context.Context, playerID string) error {
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		logger.Errorf("Failed to parse playerID: %v", err)
		return err
	}

	err = c.serviceMatch.NewMatch(ctx, playerUUID)
	if err != nil {
		logger.Errorf("Failed to new game: %v", err)
		return err
	}
	return nil
}

func (c *MatchController) Bet(ctx context.Context, playerID string, amount float64, choice string) (response dto.PlaceBetResponse, err error) {
	playerUUID, err := uuid.Parse(playerID)
	if err != nil {
		logger.Errorf("Failed to parse playerID: %v", err)
		return
	}

	number, result, err := c.serviceMatch.PlaceBet(ctx, playerUUID, amount, choice)
	if err != nil {
		logger.Errorf("Failed to place bet: %v", err)
		return
	}
	return dto.PlaceBetResponse{
		Result: result,
		Number: number,
	}, nil
}

func (c *MatchController) EndMatch(ctx context.Context, clientID string) error {
	clientUUID, err := uuid.Parse(clientID)
	if err != nil {
		logger.Errorf("Failed to parse clientID: %v", err)
		return err
	}

	err = c.serviceMatch.EndMatch(ctx, clientUUID)
	if err != nil {
		logger.Errorf("Failed to end match: %v", err)
		return err
	}
	return nil
}
