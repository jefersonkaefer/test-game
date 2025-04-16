package controller

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const ContextClientKey contextKey = "client"

var jwtSecret = []byte(os.Getenv("JWT_SECRET_KEY"))

func RequireJWT(next http.HandlerFunc) http.HandlerFunc {
	log.Default().Println("1")
	return func(w http.ResponseWriter, r *http.Request) {
		log.Default().Println("2")

		tokenStr, err := extractTokenFromRequest(r)
		log.Default().Println("3")

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			log.Default().Println("3erro;", err)
			return
		}
		log.Default().Println("4")

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			log.Default().Println("4erro;", err)
			return jwtSecret, nil
		})
		log.Default().Println("5")
		if err != nil || !token.Valid {
			log.Default().Println("5erro;", err)
			http.Error(w, "Token inválido", http.StatusUnauthorized)
			return
		}

		clientID := claims["client_id"].(string)
		ctx := context.WithValue(r.Context(), ContextClientKey, clientID)
		next(w, r.WithContext(ctx))
	}
}

func extractTokenFromRequest(r *http.Request) (string, error) {
	// Tenta pegar do Header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	// Se não vier no header, tenta na query
	if token := r.URL.Query().Get("token"); token != "" {
		return token, nil
	}

	return "", errors.New("token JWT ausente")
}
