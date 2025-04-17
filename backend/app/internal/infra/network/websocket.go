package network

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"game/api/internal/application"
	"game/api/internal/application/controller"
)

type WebSocketServer struct {
	upgrader websocket.Upgrader
	mu       sync.Mutex
	clients  map[*websocket.Conn]bool
	app      *application.App
}

type WSConfig struct {
	Upgrader websocket.Upgrader
}

func NewWebSocket(cfg WSConfig, app *application.App) *WebSocketServer {
	return &WebSocketServer{
		upgrader: cfg.Upgrader,
		clients:  make(map[*websocket.Conn]bool),
		app:      app,
	}
}

func (ws *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao atualizar para WebSocket: %v", err)
		return
	}
	defer conn.Close()
	clientID := r.Context().Value(controller.ContextClientKey).(string)
	ws.mu.Lock()
	ws.clients[conn] = true
	conn.WriteMessage(websocket.TextMessage, fmt.Appendf(nil, "Welcome, %s!", clientID))
	for user, _ := range ws.clients {
		if err := user.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s has joined the chat.", clientID))); err != nil {
			log.Println("Error sending welcome message:", err)
			return
		}
	}
	ws.mu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Erro ao ler mensagem: %v", err)
		}
		var req application.Request
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Printf("Erro ao decodificar aposta: %v", err)
			continue
		}
		ws.app.WebSocket(req)

		log.Printf("Aposta recebida: %+v", req)

		// Aqui você pode adicionar lógica para:
		// - Validar a aposta
		// - Atualizar o saldo do jogador no PostgreSQL
		// - Atualizar o cache no Redis
		// - Enviar uma resposta ao cliente
	}
}
