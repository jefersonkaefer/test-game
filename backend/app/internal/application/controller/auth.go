package controller

import (
	"context"
	"strings"

	"game/api/internal/domain/service"
	"game/api/internal/infra/logger"
	"game/api/internal/infra/session"

	"github.com/google/uuid"
)

type AuthController struct {
	authService *service.AuthService
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

func (c *AuthController) Login(ctx context.Context, username, password string) (string, error) {
	ip := ctx.Value(session.ContextKeyIP).(string)
	userAgent := ctx.Value(session.ContextKeyUserAgent).(string)

	token, err := c.authService.Login(ctx, username, password, ip, userAgent)
	if err != nil {
		logger.Errorf("Failed to login: %v", err)
		return "", err
	}

	return token, nil
}

func (c *AuthController) Logout(ctx context.Context, clientID, token string) error {
	token = strings.TrimPrefix(token, "Bearer ")
	clientIDUUID, err := uuid.Parse(clientID)
	if err != nil {
		logger.Errorf("Failed to parse clientID: %v", err)
		return err
	}
	err = c.authService.Logout(ctx, clientIDUUID, token)
	if err != nil {
		logger.Errorf("Failed to logout: %v", err)
		return err
	}

	return nil
}
