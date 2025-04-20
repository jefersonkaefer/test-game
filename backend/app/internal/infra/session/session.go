package session

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"game/api/internal/infra/logger"
)

type ContextKey string

const (
	ContextKeyIP        ContextKey = "ip"
	ContextKeyUserAgent ContextKey = "user_agent"
	ContextKeyClientID  ContextKey = "client_id"
	ContextKeySessionID ContextKey = "session_id"
	sessionKeyPrefix    string     = "session:"
)

type Session struct {
	ClientID     string    `json:"client_id"`
	IP           string    `json:"ip"`
	UserAgent    string    `json:"user_agent"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

type Manager struct {
	client    *redis.Client
	ttl       time.Duration
	jwtSecret []byte
}

func NewManager(client *redis.Client, ttl time.Duration, jwtSecret string) *Manager {
	return &Manager{
		client:    client,
		ttl:       ttl,
		jwtSecret: []byte(jwtSecret),
	}
}

func (m *Manager) ValidateJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Validating JWT token")

		tokenStr, err := extractToken(r)
		if err != nil {
			logger.Errorf("Failed to extract token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return m.jwtSecret, nil
		})

		if err != nil {
			logger.Errorf("Failed to parse token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			logger.Warn("Invalid token")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		clientID, ok := claims["client_id"].(string)
		if !ok {
			logger.Error("client_id not found in token claims")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sess, err := m.Get(r.Context(), tokenStr)
		if err != nil {
			logger.Errorf("Failed to get session: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if sess == nil {
			logger.Warn("Session not found")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		currentIP := r.RemoteAddr
		if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			currentIP = forwardedFor
		}
		currentUserAgent := r.UserAgent()

		if sess.IP != currentIP {
			logger.WithFields(logrus.Fields{
				"session_ip": sess.IP,
				"current_ip": currentIP,
			}).Warn("IP mismatch")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if sess.UserAgent != currentUserAgent {
			logger.WithFields(logrus.Fields{
				"session_user_agent": sess.UserAgent,
				"current_user_agent": currentUserAgent,
			}).Warn("User-Agent mismatch")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		err = m.UpdateActivity(r.Context(), tokenStr)
		if err != nil {
			logger.Errorf("Failed to update session activity: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.WithFields(logrus.Fields{
			"client_id": clientID,
		}).Debug("Token validated successfully")

		ctx := context.WithValue(r.Context(), ContextKeyClientID, clientID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	queryToken := r.URL.Query().Get("token")
	if queryToken != "" {
		return queryToken, nil
	}

	return "", fmt.Errorf("token not found in header or URL parameters")
}

func (m *Manager) Create(ctx context.Context, session Session) (token string, err error) {
	token, err = m.generateJWT(session.ClientID, session.IP, session.UserAgent)
	if err != nil {
		logger.Errorf("Failed to generate JWT: %v", err)
		return
	}
	logger.WithFields(logrus.Fields{
		"client_id": session.ClientID,
		"token":     token,
	}).Debug("Creating new session")

	session.CreatedAt = time.Now()
	session.LastActivity = time.Now()

	data, err := json.Marshal(session)
	if err != nil {
		logger.Errorf("Failed to marshal session: %v", err)
		return
	}
	key := sessionKeyPrefix + token
	err = m.client.Set(ctx, key, data, m.ttl).Err()
	if err != nil {
		logger.Errorf("Failed to save session: %v", err)
		return
	}

	return
}

func (m *Manager) Get(ctx context.Context, token string) (*Session, error) {
	logger.WithFields(logrus.Fields{
		"token": token,
	}).Debug("Getting session")

	key := sessionKeyPrefix + token
	data, err := m.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			logger.Debug("Session not found")
			return nil, nil
		}
		logger.Errorf("Failed to get session: %v", err)
		return nil, err
	}

	var session Session
	err = json.Unmarshal(data, &session)
	if err != nil {
		logger.Errorf("Failed to unmarshal session: %v", err)
		return nil, err
	}

	return &session, nil
}

func (m *Manager) UpdateActivity(ctx context.Context, token string) error {
	logger.WithFields(logrus.Fields{
		"token": token,
	}).Debug("Updating session activity")

	session, err := m.Get(ctx, token)
	if err != nil {
		return err
	}
	if session == nil {
		return nil
	}

	session.LastActivity = time.Now()
	data, err := json.Marshal(session)
	if err != nil {
		logger.Errorf("Failed to marshal session: %v", err)
		return err
	}

	key := sessionKeyPrefix + token
	err = m.client.Set(ctx, key, data, m.ttl).Err()
	if err != nil {
		logger.Errorf("Failed to update session: %v", err)
		return err
	}

	return nil
}

func (m *Manager) Delete(ctx context.Context, token string) error {
	logger.WithFields(logrus.Fields{
		"token": token,
	}).Debug("Deleting session")

	key := sessionKeyPrefix + token
	err := m.client.Del(ctx, key).Err()
	if err != nil {
		logger.Errorf("Failed to delete session: %v", err)
		return err
	}

	return nil
}

func (m *Manager) DeleteAllForClient(ctx context.Context, clientID string) error {
	logger.WithFields(logrus.Fields{
		"client_id": clientID,
	}).Debug("Deleting all sessions for client")

	pattern := sessionKeyPrefix + "*"
	keys, err := m.client.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Errorf("Failed to get session keys: %v", err)
		return err
	}

	for _, key := range keys {
		session, err := m.Get(ctx, key[len(sessionKeyPrefix):]) // Remove o prefixo "session:"
		if err != nil {
			continue
		}
		if session != nil && session.ClientID == clientID {
			err = m.Delete(ctx, key[len(sessionKeyPrefix):])
			if err != nil {
				logger.Errorf("Failed to delete session %s: %v", key, err)
			}
		}
	}

	return nil
}

func (m *Manager) DeleteAllSessionsForClient(ctx context.Context, clientID string) error {
	logger.WithFields(logrus.Fields{
		"client_id": clientID,
	}).Debug("Deleting all sessions for client")

	pattern := sessionKeyPrefix + clientID + "*"
	keys, err := m.client.Keys(ctx, pattern).Result()
	if err != nil {
		logger.Errorf("Failed to get session keys: %v", err)
		return err
	}

	for _, key := range keys {
		err = m.Delete(ctx, key[len(sessionKeyPrefix):])
		if err != nil {
			logger.Errorf("Failed to delete session %s: %v", key, err)
		}
	}

	return nil
}

func GenerateID() string {
	return uuid.New().String()
}

func (m *Manager) generateJWT(clientID, ip, userAgent string) (string, error) {

	claims := jwt.MapClaims{
		"client_id":  clientID,
		"ip":         ip,
		"user_agent": userAgent,
		"iss":        "game-api",
		"exp":        time.Now().Add(24 * time.Hour).Unix(),
		"iat":        time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(m.jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
