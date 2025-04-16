package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	cache "github.com/redis/go-redis/v9"

	"game/api/internal/game"
	"game/api/internal/game/entity"
	"game/api/internal/network"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic recovery: %v", err)
		}
	}()
	/*
		conn, err := dbConn()
		if err != nil {
			log.Default().Fatal("ERROR occurs: %v", err)
		}

		pg := database.NewPostgres(conn)
		cacheClient := cacheConn()
		redis := database.NewRedis(cacheClient)
	*/
	wsCfg := network.WSConfig{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Em produção, implemente uma verificação adequada
			},
		},
	}
	ws := network.NewWebSocket(wsCfg)
	http.HandleFunc("/ws", ws.HandleConnections)

	log.Println("Servidor WebSocket iniciado na porta :8080")
	err := http.ListenAndServe(":4300", nil)
	if err != nil {
		log.Fatalf("Erro ao iniciar o servidor: %v", err)
	}
	p := entity.NewPlayer()
	cfg := game.Config{
		MaxNumberDraw: 10,
	}
	g := game.NewGame(cfg)
	m := game.NewMatch(g)
	mr, err := m.Play(p, 10, game.Even)

	if err != nil {
		log.Println("ERRO:", err.Error())
	}
	fmt.Printf("%v", mr)
}

func dbConn() (*sql.DB, error) {
	host := "db"
	port := "5432"
	user := "root"
	pw := "t3st"
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
