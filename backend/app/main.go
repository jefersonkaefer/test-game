package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
	cache "github.com/redis/go-redis/v9"

	"game/api/internal/controller"
	"game/api/internal/infra/database"
	"game/api/internal/infra/network"
	"game/api/internal/repository"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic recovery: %v", err)
		}
	}()

	conn, err := dbConn()
	if err != nil {
		log.Default().Fatal("ERROR occurs: %v", err)
	}

	pg := database.NewPostgres(conn)
	cacheClient := cacheConn()
	redis := database.NewRedis(cacheClient)

	wsCfg := network.WSConfig{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Em produção, implemente uma verificação adequada
			},
		},
	}
	clientsRepo := repository.NewClient(pg, redis)
	app := controller.NewApp(clientsRepo)
	ws := network.NewWebSocket(wsCfg, app)
	api := network.NewWebServer(app)

	http.HandleFunc("/ws", controller.RequireJWT(ws.HandleConnections))

	go func() {
		log.Println("Servidor WebSocket iniciado na porta :4300")
		if err := http.ListenAndServe(":4300", nil); err != nil {
			log.Fatalf("Erro ao iniciar o servidor WebSocket: %v", err)
		}
	}()

	http.HandleFunc("/client", api.NewClient)
	http.HandleFunc("/login", api.Login)
	// Inicia o servidor na porta 8080
	go func() {
		log.Println("Servidor iniciado na porta :8000")
		if err := http.ListenAndServe(":8000", nil); err != nil {
			log.Fatalf("Erro ao iniciar o servidor HTTP: %v", err)
		}
	}()
	if err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
	select {}
}

func dbConn() (*sql.DB, error) {
	host := "db"
	port := "5432"
	user := "postgres"
	pw := os.Getenv("POSTGRES_PASSWORD")
	dbName := "game"
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, pw, dbName)
	return sql.Open("postgres", connStr)
}

func cacheConn() *cache.Client {
	addr := "redis:6379"
	password := "t3st"
	return cache.NewClient(&cache.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
}
