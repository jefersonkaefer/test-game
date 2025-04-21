package network

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"game/api/internal/application/controller"
	"game/api/internal/application/dto"
	"game/api/internal/errs"
	"game/api/internal/infra/logger"
	"game/api/internal/infra/session"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
)

const (
	ActionNewMatch string = "new_match"
	ActionPlaceBet string = "place_bet"
	ActionWallet   string = "wallet"
	ActionEndMatch string = "end_match"
	pingPeriod            = 30 * time.Second
	pongWait              = 60 * time.Second
	writeWait             = 10 * time.Second
)

type WSResponse struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
	Error  string      `json:"error,omitempty"`
}

type WebSocketRequest struct {
	Action string      `json:"action"`
	Data   interface{} `json:"data"`
}

type Client struct {
	Conn *websocket.Conn
	Send chan []byte
}

type WebServer struct {
	*chi.Mux
	clientController *controller.ClientController
	authController   *controller.AuthController
	matchController  *controller.MatchController
	upgrader         websocket.Upgrader
	clients          map[string]*Client
	sessionManager   *session.Manager
}

func NewWebServer(
	clientController *controller.ClientController,
	authController *controller.AuthController,
	matchController *controller.MatchController,
	sessionManager *session.Manager,
) *WebServer {
	ws := &WebServer{
		Mux:              chi.NewMux(),
		clientController: clientController,
		authController:   authController,
		matchController:  matchController,
		sessionManager:   sessionManager,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[string]*Client),
	}

	ws.setupRoutes()
	return ws
}

func (ws *WebServer) setupRoutes() {
	ws.Post("/register", ws.register)
	ws.Post("/login", ws.login)
	ws.Post("/logout", ws.sessionManager.ValidateJWT(ws.logout))
	ws.Get("/wallet", ws.sessionManager.ValidateJWT(ws.wallet))
	ws.Get("/ws", ws.sessionManager.ValidateJWT(ws.handleWebSocket))
}

func (ws *WebServer) register(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	res, err := ws.clientController.Create(r.Context(), req.Username, req.Password)
	if err != nil {
		if err == errs.ErrUsernameExists {
			http.Error(w, "Username already exists", http.StatusConflict)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (ws *WebServer) login(w http.ResponseWriter, r *http.Request) {
	var req dto.ClientLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ip := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = forwardedFor
	}
	userAgent := r.UserAgent()

	ctx := context.WithValue(r.Context(), session.ContextKeyIP, ip)
	ctx = context.WithValue(ctx, session.ContextKeyUserAgent, userAgent)

	token, err := ws.authController.Login(ctx, req.Username, req.Password)
	if err != nil {
		switch err {
		case errs.ErrNotFound:
			http.Error(w, "Client not found", http.StatusNotFound)
		case errs.ErrInvalidPassword:
			http.Error(w, "Invalid password", http.StatusUnauthorized)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func (ws *WebServer) logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	clientID, ok := r.Context().Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		http.Error(w, "client ID is required", http.StatusUnauthorized)
		return
	}
	ws.authController.Logout(r.Context(), clientID, token)
	w.WriteHeader(http.StatusOK)
}

func (ws *WebServer) wallet(w http.ResponseWriter, r *http.Request) {
	clientID, ok := r.Context().Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		http.Error(w, "client ID is required", http.StatusUnauthorized)
		return
	}
	balance, err := ws.clientController.GetBalance(r.Context(), clientID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(balance)
}

func (ws *WebServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	clientID, ok := r.Context().Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		http.Error(w, "client ID is required", http.StatusUnauthorized)
		return
	}

	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		Conn: conn,
		Send: make(chan []byte, 256),
	}

	ws.clients[clientID] = client

	ctx := context.WithValue(r.Context(), session.ContextKeyClientID, clientID)

	go ws.handleConnection(ctx, conn)
}

func (ws *WebServer) handleConnection(ctx context.Context, conn *websocket.Conn) {
	clientID, _ := ctx.Value(session.ContextKeyClientID).(string)
	logger.Infof("New WebSocket connection established for client %s", clientID)

	defer func() {
		logger.Infof("Closing WebSocket connection for client %s", clientID)
		conn.Close()
		if clientID != "" {
			delete(ws.clients, clientID)
		}
	}()

	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	conn.SetPingHandler(func(string) error {
		return conn.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(writeWait))
	})

	conn.SetReadDeadline(time.Now().Add(pongWait))

	go func() {
		pingCtx, cancel := context.WithCancel(context.Background())
		defer cancel()

		ticker := time.NewTicker(pingPeriod)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(writeWait)); err != nil {
					logger.Errorf("Error sending ping: %v", err)
					return
				}
			case <-pingCtx.Done():
				return
			}
		}
	}()

	for {
		var request WebSocketRequest
		err := conn.ReadJSON(&request)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Errorf("Error reading message: %v", err)
			}
			break
		}

		conn.SetReadDeadline(time.Now().Add(pongWait))

		response := ws.handleRequest(ctx, request)
		if err := conn.WriteJSON(response); err != nil {
			logger.Errorf("Error writing response: %v", err)
			break
		}
	}
}

func (ws *WebServer) handleRequest(ctx context.Context, request WebSocketRequest) *WSResponse {
	clientID, _ := ctx.Value(session.ContextKeyClientID).(string)
	logger.Infof("Handling request from client %s: %s", clientID, request.Action)

	msgCtx := context.Background()
	msgCtx = context.WithValue(msgCtx, session.ContextKeyClientID, clientID)

	msgCtx, cancel := context.WithTimeout(msgCtx, 1000*time.Second)
	defer cancel()

	var response *WSResponse

	switch request.Action {
	case ActionNewMatch:
		response = ws.handleNewMatch(msgCtx)
	case ActionPlaceBet:
		response = ws.handleBet(msgCtx, request.Data)
	case ActionWallet:
		response = ws.handleWallet(msgCtx)
	case ActionEndMatch:
		response = ws.handleEndMatch(msgCtx)
	default:
		logger.Errorf("Invalid action from client %s: %s", clientID, request.Action)
		response = ws.errorResponse("Invalid action")
	}

	return response
}

func (ws *WebServer) handleNewMatch(ctx context.Context) *WSResponse {
	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return ws.errorResponse("client ID is required")
	}

	err := ws.matchController.NewMatch(ctx, clientID)
	if err != nil {
		logger.Errorf("Failed to new match: %v", err)
		return ws.errorResponse(err.Error())
	}
	return ws.successResponse(ActionNewMatch, nil)
}

func (ws *WebServer) handleBet(ctx context.Context, body interface{}) *WSResponse {
	var req struct {
		Amount float64 `json:"amount"`
		Choice string  `json:"choice"`
	}
	if err := ws.unmarshalRequest(body, &req); err != nil {
		logger.Errorf("Error unmarshaling bet request: %v", err)
		return ws.errorResponse(err.Error())
	}

	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		logger.Errorf("Invalid client ID in context")
		return ws.errorResponse("client ID is required")
	}

	result, err := ws.matchController.Bet(ctx, clientID, req.Amount, req.Choice)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			logger.Errorf("Bet operation canceled for client %s: %v", clientID, err)
			return ws.errorResponse("operation timed out, please try again")
		}
		logger.Errorf("Error placing bet for client %s: %v", clientID, err)
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse(ActionPlaceBet, result)
}

func (ws *WebServer) handleWallet(ctx context.Context) *WSResponse {
	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return ws.errorResponse("client ID is required")
	}

	wallet, err := ws.clientController.GetBalance(ctx, clientID)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse(ActionWallet, wallet)
}

func (ws *WebServer) handleEndMatch(ctx context.Context) *WSResponse {
	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return ws.errorResponse("client ID is required")
	}

	err := ws.matchController.EndMatch(ctx, clientID)
	if err != nil {
		return ws.errorResponse(err.Error())
	}
	return ws.successResponse(ActionEndMatch, nil)
}

func (ws *WebServer) unmarshalRequest(body interface{}, req interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	if err := json.Unmarshal(data, req); err != nil {
		return fmt.Errorf("failed to unmarshal request body: %v", err)
	}

	return nil
}

func (ws *WebServer) errorResponse(msg string) *WSResponse {
	return &WSResponse{
		Action: "error",
		Error:  msg,
	}
}

func (ws *WebServer) successResponse(action string, data interface{}) *WSResponse {
	return &WSResponse{
		Action: action,
		Data:   data,
	}
}

func (ws *WebServer) Start(port int) error {
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      ws,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Infof("Server starting on port %d", port)
	return server.ListenAndServe()
}

func (ws *WebServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws.Mux.ServeHTTP(w, r)
}
