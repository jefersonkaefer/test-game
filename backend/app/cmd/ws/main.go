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
	logger.Info("Starting WebSocket server")

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
	ws := network.NewWebSocket(network.WSConfig{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}, app)

	http.HandleFunc("/ws", controller.RequireJWT(ws.HandleConnections))

	logger.Info("WebSocket server listening on :4300")
	if err := http.ListenAndServe(":4300", nil); err != nil {
		log.Fatalf("Failed to start WebSocket server: %v", err)
	}
}
