package main

import (
	"dicegame/game"
	"dicegame/handlers"
	"dicegame/middleware"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Inicializa o gerenciador de jogos
	gameManager := game.NewGameManager()
	authHandler := handlers.NewAuthHandler()

	// Configura o roteador
	r := mux.NewRouter()

	// Middleware para todas as rotas
	r.Use(middleware.JSONResponse)

	// Rotas públicas
	authHandler.RegisterRoutes(r)

	// Rotas protegidas
	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware)

	// WebSocket com token na URL
	r.HandleFunc("/ws/{token}", gameManager.HandleConnections)

	// Inicia o servidor
	log.Println("Servidor iniciado na porta 8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Erro ao iniciar servidor: ", err)
	}
} 