package controller

import (
	"context"

	"game/api/internal/domain/service"
	"game/api/internal/infra/logger"
	"game/api/internal/infra/session"
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

func (c *AuthController) Logout(ctx context.Context, token string) error {
	err := c.authService.Logout(ctx, token)
	if err != nil {
		logger.Errorf("Failed to logout: %v", err)
		return err
	}

	return nil
}
