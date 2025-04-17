package network

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"

	"game/api/internal/application"
	"game/api/internal/application/controller"
	"game/api/internal/infra/logger"
)

type WebSocketServer struct {
	upgrader websocket.Upgrader
	mu       sync.Mutex
	clients  map[*websocket.Conn]bool
	app      *application.App
}

type WSConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
}

func NewWebSocket(cfg WSConfig, app *application.App) *WebSocketServer {
	logger.Info("Initializing WebSocket server")
	return &WebSocketServer{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.ReadBufferSize,
			WriteBufferSize: cfg.WriteBufferSize,
		},
		clients: make(map[*websocket.Conn]bool),
		app:     app,
	}
}

func (ws *WebSocketServer) HandleConnections(w http.ResponseWriter, r *http.Request) {
	logger.Info("New WebSocket connection attempt")

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	logger.Info("WebSocket connection established")

	// Verificação segura do clientID
	clientIDValue := r.Context().Value(controller.ContextClientKey)
	if clientIDValue == nil {
		logger.Error("Client ID not found in context")
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Unauthorized"))
		return
	}

	clientID, ok := clientIDValue.(string)
	if !ok {
		logger.Error("Invalid client ID type in context")
		conn.WriteMessage(websocket.TextMessage, []byte("Error: Invalid client ID"))
		return
	}

	ws.mu.Lock()
	ws.clients[conn] = true
	conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Welcome, %s!", clientID)))
	for user := range ws.clients {
		if err := user.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("%s has joined the chat.", clientID))); err != nil {
			logger.Errorf("Error sending welcome message: %v", err)
			continue
		}
	}
	ws.mu.Unlock()

	for {
		var req application.Request
		err := conn.ReadJSON(&req)
		if err != nil {
			logger.Errorf("Error reading message: %v", err)
			break
		}

		logger.WithFields(logrus.Fields{
			"action": req.Action,
		}).Debug("Received WebSocket message")

		res := ws.app.WebSocket(req)

		err = conn.WriteJSON(res)
		if err != nil {
			logger.Errorf("Error writing message: %v", err)
			break
		}
	}

	logger.Info("WebSocket connection closed")
}
