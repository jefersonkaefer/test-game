package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"game/api/internal/application"
	"game/api/internal/application/controller"
	"game/api/internal/application/repository"
	"game/api/internal/domain/service"
	"game/api/internal/infra/network"
	"game/api/internal/infra/session"
)

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
	redis := application.RedisConn(ctx)
	defer redis.Close()

	sessionManager := session.NewManager(redis, 24*time.Hour, os.Getenv("JWT_SECRET_KEY"))

	clientsRepo := repository.NewClient(db, redis)
	matchRepo := repository.NewMatch(db, redis)
	walletRepo := repository.NewWallet(db, redis)

	clientsService := service.NewClientService(clientsRepo, walletRepo)
	matchService := service.NewMatchService(matchRepo, clientsService)
	authService := service.NewAuthService(clientsService, sessionManager)

	clientsCtrl := controller.NewClientController(clientsService)
	authCtrl := controller.NewAuthController(authService)
	matchCtrl := controller.NewMatchController(matchService)

	wsConfig := network.WSConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		ReadTimeout:     60 * time.Second,
		WriteTimeout:    60 * time.Second,
	}

	api := network.NewWebServer(wsConfig, sessionManager, clientsCtrl, matchCtrl, authCtrl)

	mux := http.NewServeMux()
	mux.Handle("/", api)

	log.Println("HTTP server started on port :8000")
	if err := http.ListenAndServe(":8000", mux); err != nil {
		log.Fatalf("An error occurred while starting HTTP: %v", err)
	}
}
