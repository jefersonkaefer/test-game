package service

import (
	"context"

	"game/api/internal/errs"
	"game/api/internal/infra/logger"
	"game/api/internal/infra/session"

	"github.com/google/uuid"
)

type AuthService struct {
	clientService  *ClientService
	sessionManager *session.Manager
}

func NewAuthService(clientService *ClientService, sessionManager *session.Manager) *AuthService {
	return &AuthService{
		clientService:  clientService,
		sessionManager: sessionManager,
	}
}

func (s *AuthService) Login(ctx context.Context, username, password, ip, userAgent string) (string, error) {
	client, err := s.clientService.GetByUsername(ctx, username)
	if err != nil {
		return "", err
	}

	if !client.CheckPasswordHash(password) {
		return "", errs.ErrInvalidPassword
	}

	sess := session.Session{
		ClientID:  client.GetID().String(),
		IP:        ip,
		UserAgent: userAgent,
	}

	token, err := s.sessionManager.Create(ctx, sess)
	if err != nil {
		logger.Errorf("Failed to create session: %v", err)
		return "", err
	}

	return token, nil
}

func (s *AuthService) Logout(ctx context.Context, clientID uuid.UUID, token string) error {
	err := s.clientService.RefreshWallet(ctx, clientID)
	if err != nil {
		logger.Errorf("Failed to refresh wallet: %v", err)
		return err
	}
	err = s.sessionManager.Delete(ctx, token)
	if err != nil {
		logger.Errorf("Failed to delete session: %v", err)
		return err
	}
	return nil
}
