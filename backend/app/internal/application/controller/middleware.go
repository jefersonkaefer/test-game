package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"

	"game/api/internal/infra/logger"
)

type contextKey string

const ContextClientKey contextKey = "client_id"

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func RequireJWT(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Debug("Validating JWT token")

		// Tentar extrair o token do cabeçalho ou dos parâmetros da URL
		tokenStr, err := extractToken(r)
		if err != nil {
			logger.Errorf("Failed to extract token: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims := &jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
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

		logger.WithFields(logrus.Fields{
			"subject": claims.Subject,
		}).Debug("Token validated successfully")

		ctx := context.WithValue(r.Context(), ContextClientKey, claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func extractToken(r *http.Request) (string, error) {
	// Tentar extrair o token do cabeçalho Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			return parts[1], nil
		}
	}

	// Tentar extrair o token dos parâmetros da URL
	queryToken := r.URL.Query().Get("token")
	if queryToken != "" {
		return queryToken, nil
	}

	return "", fmt.Errorf("token not found in header or URL parameters")
}
