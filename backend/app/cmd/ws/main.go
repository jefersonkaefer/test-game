package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	"game/api/internal/application"
	"game/api/internal/application/controller"
	"game/api/internal/application/repository"
	"game/api/internal/infra/database"
	"game/api/internal/infra/network"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic recovery: %v", err)
		}
	}()

	conn, err := application.DbConn()
	if err != nil {
		log.Fatalf("ERROR occurs: %v", err)
	}

	pg := database.NewPostgres(conn)
	cacheClient := application.CacheConn()
	redis := database.NewRedis(cacheClient)

	wsCfg := network.WSConfig{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	clientsRepo := repository.NewClient(pg, redis)
	clientCtrl := controller.NewClient(clientsRepo)

	app := &application.App{
		ClientCtrl: clientCtrl,
	}
	ws := network.NewWebSocket(wsCfg, app)

	http.HandleFunc("/ws", controller.RequireJWT(ws.HandleConnections))

	log.Println("WebSocket server started on port :4300")
	if err := http.ListenAndServe(":4300", nil); err != nil {
		log.Fatalf("Erro ao iniciar o servidor WebSocket: %v", err)
	}
}
