package network

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"

	"game/api/internal/game"
)

type WebSocketServer struct {
	upgrader websocket.Upgrader
	mu       sync.Mutex
	clients  map[*websocket.Conn]bool
}
type WSConfig struct {
	Upgrader websocket.Upgrader
}

func NewWebSocket(cfg WSConfig) *WebSocketServer {
	return &WebSocketServer{
		upgrader: cfg.Upgrader,
		clients:  make(map[*websocket.Conn]bool),
	}
}

func (ws *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Erro ao atualizar para WebSocket: %v", err)
		return
	}
	defer conn.Close()

	ws.mu.Lock()
	ws.clients[conn] = true
	ws.mu.Unlock()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Erro ao ler mensagem: %v", err)
			ws.mu.Lock()
			delete(ws.clients, conn)
			ws.mu.Unlock()
			break
		}

		var bet game.BetType
		if err := json.Unmarshal(msg, &bet); err != nil {
			log.Printf("Erro ao decodificar aposta: %v", err)
			continue
		}

		log.Printf("Aposta recebida: %+v", bet)

		// Aqui você pode adicionar lógica para:
		// - Validar a aposta
		// - Atualizar o saldo do jogador no PostgreSQL
		// - Atualizar o cache no Redis
		// - Enviar uma resposta ao cliente
	}
}
