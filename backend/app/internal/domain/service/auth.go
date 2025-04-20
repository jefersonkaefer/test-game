package service

import (
	"context"

	"game/api/internal/errs"
	"game/api/internal/infra/logger"
	"game/api/internal/infra/session"
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

func (s *AuthService) Logout(ctx context.Context, token string) error {
	return s.sessionManager.Delete(ctx, token)
}

func (s *AuthService) LogoutAll(ctx context.Context, clientID string) error {
	return s.sessionManager.DeleteAllForClient(ctx, clientID)
}

func (s *AuthService) LogoutAllSessionsForClient(ctx context.Context, clientID string) error {
	return s.sessionManager.DeleteAllSessionsForClient(ctx, clientID)
}
