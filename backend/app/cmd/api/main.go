package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"game/api/internal/application"
	"game/api/internal/application/controller"
	"game/api/internal/application/repository"
	"game/api/internal/domain/service"
	"game/api/internal/infra/database"
	"game/api/internal/infra/network"
	"game/api/internal/infra/session"
)

func validateJWTSecret() (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET_KEY environment variable is not set")
	}
	return jwtSecret, nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic recovery: %v", err)
		}
	}()

	ctx := context.Background()
	db, err := application.DbConn(ctx)
	if err != nil {
		log.Fatalf("ERROR occurs: %v", err)
	}
	defer db.Close()
	redisConn := application.RedisConn(ctx)
	defer redisConn.Close()

	redis := database.NewRedis(redisConn)

	jwtSecret, err := validateJWTSecret()
	if err != nil {
		log.Fatalf("ERROR validating JWT secret: %v", err)
	}

	sessionManager := session.NewManager(redisConn, 24*time.Hour, jwtSecret)

	clientsRepo := repository.NewClients(redis, db)
	walletRepo := repository.NewWallets(redis, db)
	playerRepo := repository.NewPlayers(redis, clientsRepo, walletRepo)

	clientsService := service.NewClientService(clientsRepo, walletRepo)
	matchService := service.NewMatchService(playerRepo, walletRepo)
	authService := service.NewAuthService(clientsService, sessionManager)

	clientsCtrl := controller.NewClientController(clientsService)
	authCtrl := controller.NewAuthController(authService)
	matchCtrl := controller.NewMatchController(matchService)

	api := network.NewWebServer(clientsCtrl, authCtrl, matchCtrl, sessionManager)
	mux := http.NewServeMux()
	mux.Handle("/", api)

	server := &http.Server{
		Addr:    ":8000",
		Handler: mux,
	}

	//canal para receber sinais do o.s
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	//inicia o servidor em background
	go func() {
		log.Println("HTTP server started on port :8000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("An error occurred while starting HTTP: %v", err)
		}
	}()

	//sinal de interrupção
	<-stop
	log.Println("Shutting down server...")

	//timeout para o shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	//graceful shutdown
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
