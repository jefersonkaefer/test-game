package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"game/api/internal/application/controller"
	"game/api/internal/application/dto"
	"game/api/internal/errs"
	"game/api/internal/infra/logger"
	"game/api/internal/infra/session"

	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

const (
	ActionCreateMatch  string = "create_match"
	ActionJoinMatch    string = "join_match"
	ActionLeaveMatch   string = "leave_match"
	ActionPlaceBet     string = "place_bet"
	ActionChooseParity string = "choose_parity"
	ActionGetMatch     string = "get_match"
	ActionStartMatch   string = "start_match"
	ActionEndMatch     string = "end_match"
)

type WSConfig struct {
	ReadBufferSize  int
	WriteBufferSize int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

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

type webServer struct {
	clientCtrl *controller.ClientController
	matchCtrl  *controller.MatchController
	authCtrl   *controller.AuthController
	sessionMgr *session.Manager
	upgrader   websocket.Upgrader
	mu         sync.Mutex
	clients    map[string]*Client
}

func NewWebServer(
	cfg WSConfig,
	sessionMgr *session.Manager,
	clientCtrl *controller.ClientController,
	matchCtrl *controller.MatchController,
	authCtrl *controller.AuthController,
) *webServer {
	return &webServer{
		clientCtrl: clientCtrl,
		matchCtrl:  matchCtrl,
		sessionMgr: sessionMgr,
		authCtrl:   authCtrl,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  cfg.ReadBufferSize,
			WriteBufferSize: cfg.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[string]*Client),
	}
}

func (ws *webServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Rotas públicas
	switch r.URL.Path {
	case "/client":
		ws.NewClient(w, r)
	case "/login":
		ws.Authenticate(w, r)
	default:
		protected := ws.sessionMgr.ValidateJWT(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch r.URL.Path {
			case "/ws":
				ws.handleConnection(w, r)
			default:
				http.NotFound(w, r)
			}
		}))
		protected.ServeHTTP(w, r)
	}
}

func (ws *webServer) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("Failed to upgrade connection", logrus.Fields{"error": err})
		return
	}

	sessionID := session.GenerateID()
	client := &Client{
		Conn: conn,
	}

	ws.clients[sessionID] = client
	if err := conn.WriteJSON(WSResponse{
		Action: "connection",
		Data: map[string]interface{}{
			"sessionID": sessionID,
		},
	}); err != nil {
		logger.Error("Failed to send connection message", logrus.Fields{"error": err})
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logger.Error("WebSocket read error", logrus.Fields{"error": err})
			}
			break
		}

		ctx := context.WithValue(r.Context(), session.ContextKeySessionID, sessionID)

		var request WebSocketRequest
		if err := json.Unmarshal(message, &request); err != nil {
			logger.Error("Failed to parse WebSocket message", logrus.Fields{"error": err})
			continue
		}

		ws.handleRequest(ctx, conn, request)
	}

	ws.mu.Lock()
	delete(ws.clients, sessionID)
	ws.mu.Unlock()
}

type SuccessResponse struct {
	Status    string      `json:"status"`
	Code      int         `json:"code"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

type ErrorResponse struct {
	Status    string      `json:"status"`
	Code      int         `json:"code"`
	Error     string      `json:"error"`
	Message   string      `json:"message,omitempty"`
	Details   interface{} `json:"details,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

func JSONSuccess(w http.ResponseWriter, code int, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	resp := SuccessResponse{
		Status:    "success",
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}
	json.NewEncoder(w).Encode(resp)
}

func JSONError(w http.ResponseWriter, code int, errCode, message string, details interface{}) {
	w.Header().Set("Content-Type", "application/json")
	resp := ErrorResponse{
		Status:    "error",
		Code:      code,
		Error:     errCode,
		Message:   message,
		Details:   details,
		Timestamp: time.Now().UTC(),
	}
	json.NewEncoder(w).Encode(resp)
}

func (ws *webServer) NewClient(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateClientRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	res, err := ws.clientCtrl.Create(r.Context(), req.Username, req.Password)
	if err != nil {
		if err == errs.ErrUsernameExists {
			JSONError(w, http.StatusConflict,
				"USERNAME_ALREADY_EXISTS",
				"This username already exists",
				nil,
			)
			return
		}
		JSONError(w, http.StatusInternalServerError,
			http.StatusText(http.StatusInternalServerError),
			http.StatusText(http.StatusInternalServerError),
			nil,
		)
		return
	}
	JSONSuccess(w, http.StatusCreated,
		"Client created successfully.",
		res,
	)
}

func (ws *webServer) Authenticate(w http.ResponseWriter, r *http.Request) {
	var req dto.ClientLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	ip := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		ip = forwardedFor
	}
	userAgent := r.UserAgent()

	ctx := context.WithValue(r.Context(), session.ContextKeyIP, ip)
	ctx = context.WithValue(ctx, session.ContextKeyUserAgent, userAgent)

	res, err := ws.authCtrl.Login(ctx, req.Username, req.Password)
	if err != nil {
		switch err {
		case errs.ErrNotFound:
			JSONError(w, http.StatusNotFound, "NOT_FOUND", "Client not found", nil)
		case errs.ErrInvalidPassword:
			JSONError(w, http.StatusUnauthorized, "INVALID_PASSWORD", "Invalid password", nil)
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		}
		return
	}

	JSONSuccess(w, http.StatusOK, "Login successful", res)
}

func (ws *webServer) handleRequest(ctx context.Context, conn *websocket.Conn, request WebSocketRequest) {
	var response *WSResponse

	switch request.Action {
	case ActionCreateMatch:
		response = ws.handleCreateMatch(ctx, request.Data)
	case ActionJoinMatch:
		response = ws.handleJoinMatch(ctx, request.Data)
	case ActionLeaveMatch:
		response = ws.handleLeaveMatch(ctx, request.Data)
	case ActionPlaceBet:
		response = ws.handlePlaceBet(ctx, request.Data)
	case ActionGetMatch:
		response = ws.handleGetMatch(ctx, request.Data)
	case ActionStartMatch:
		response = ws.handleStartMatch(ctx, request.Data)
	case ActionEndMatch:
		response = ws.handleEndMatch(ctx, request.Data)
	default:
		response = &WSResponse{
			Action: "error",
			Error:  fmt.Sprintf("invalid action: %s", request.Action),
		}
	}

	if response != nil {
		conn.WriteJSON(response)
	}
}

func (ws *webServer) handleCreateMatch(ctx context.Context, body interface{}) *WSResponse {
	var req dto.CreateMatchRequest
	if err := ws.unmarshalRequest(body, &req); err != nil {
		return ws.errorResponse("invalid request format: " + err.Error())
	}

	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return ws.errorResponse(errs.ErrInvalidToken.Error())
	}

	match, err := ws.matchCtrl.CreateMatch(ctx, clientID, req)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse("match_created", match)
}

func (ws *webServer) handleJoinMatch(ctx context.Context, body interface{}) *WSResponse {
	var req dto.AddPlayerRequest
	if err := ws.unmarshalRequest(body, &req); err != nil {
		return ws.errorResponse("invalid request format: " + err.Error())
	}

	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return ws.errorResponse(errs.ErrInvalidToken.Error())
	}

	res, err := ws.matchCtrl.JoinMatch(ctx, req.MatchID, clientID)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse("match_joined", res)
}

func (ws *webServer) handleLeaveMatch(ctx context.Context, body interface{}) *WSResponse {
	var req dto.AddPlayerRequest
	if err := ws.unmarshalRequest(body, &req); err != nil {
		return ws.errorResponse("invalid request format: " + err.Error())
	}

	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return ws.errorResponse(errs.ErrInvalidToken.Error())
	}

	err := ws.matchCtrl.LeaveMatch(ctx, req)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse("match_left", nil)
}

func (ws *webServer) handlePlaceBet(ctx context.Context, body interface{}) *WSResponse {
	var req dto.CreateBetRequest
	if err := ws.unmarshalRequest(body, &req); err != nil {
		return ws.errorResponse("invalid request format: " + err.Error())
	}

	clientID, ok := ctx.Value(session.ContextKeyClientID).(string)
	if !ok || clientID == "" {
		return ws.errorResponse(errs.ErrInvalidToken.Error())
	}

	err := ws.matchCtrl.PlaceBetAndChoose(ctx, clientID, req)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse("bet_placed", nil)
}

func (ws *webServer) handleGetMatch(ctx context.Context, body interface{}) *WSResponse {
	var reqData struct {
		MatchID string `json:"match_id"`
	}
	if err := ws.unmarshalRequest(body, &reqData); err != nil {
		return ws.errorResponse(err.Error())
	}

	match, err := ws.matchCtrl.GetMatch(ctx, reqData.MatchID)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse(ActionGetMatch, match)
}

func (ws *webServer) handleStartMatch(ctx context.Context, body interface{}) *WSResponse {
	var reqData struct {
		MatchID string `json:"match_id"`
	}
	if err := ws.unmarshalRequest(body, &reqData); err != nil {
		return ws.errorResponse(err.Error())
	}

	err := ws.matchCtrl.StartMatch(ctx, reqData.MatchID)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse("match_started", nil)
}

func (ws *webServer) handleEndMatch(ctx context.Context, body interface{}) *WSResponse {
	var reqData struct {
		MatchID string `json:"match_id"`
	}
	if err := ws.unmarshalRequest(body, &reqData); err != nil {
		return ws.errorResponse(err.Error())
	}

	err := ws.matchCtrl.EndMatch(ctx, reqData.MatchID)
	if err != nil {
		return ws.errorResponse(err.Error())
	}

	return ws.successResponse("match_ended", nil)
}

func (ws *webServer) unmarshalRequest(body interface{}, req interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	if err := json.Unmarshal(data, req); err != nil {
		return fmt.Errorf("failed to unmarshal request body: %v", err)
	}

	return nil
}

func (ws *webServer) errorResponse(msg string) *WSResponse {
	return &WSResponse{
		Action: "error",
		Error:  msg,
	}
}

func (ws *webServer) successResponse(action string, data interface{}) *WSResponse {
	return &WSResponse{
		Action: action,
		Data:   data,
	}
}
