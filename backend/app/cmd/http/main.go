package main

import (
	"log"
	"net/http"

	"game/api/internal/application"
	"game/api/internal/application/controller"
	"game/api/internal/application/repository"
	"game/api/internal/infra/logger"
	"game/api/internal/infra/network"
)

func main() {
	logger.Info("Starting HTTP server")

	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("Panic recovered: %v", r)
		}
	}()

	db, err := application.DbConn()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	cache := application.CacheConn()
	defer cache.Close()

	clientRepo := repository.NewClient(db, cache)
	clientCtrl := controller.NewClient(clientRepo)

	app := application.NewApp(clientCtrl)

	server := network.NewWebServer(app)

	http.HandleFunc("/client", server.NewClient)
	http.HandleFunc("/login", server.Login)

	logger.Info("HTTP server listening on :8000")
	if err := http.ListenAndServe(":8000", nil); err != nil {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}
