package network

import (
	"encoding/json"
	"game/api/internal/application"
	"game/api/internal/infra/logger"
	"net/http"
	"time"

	"game/api/internal/application/controller"

	"github.com/sirupsen/logrus"
)

type webServer struct {
	app *application.App
}

func NewWebServer(app *application.App) *webServer {
	logger.Info("Initializing web server")
	return &webServer{app: app}
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
	logger.WithFields(logrus.Fields{
		"code":    code,
		"message": message,
	}).Debug("Sending success response")

	response := map[string]interface{}{
		"success": true,
		"message": message,
		"data":    data,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func JSONError(w http.ResponseWriter, code int, errCode, message string, details interface{}) {
	logger.WithFields(logrus.Fields{
		"code":    code,
		"errCode": errCode,
		"message": message,
	}).Error("Sending error response")

	response := map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    errCode,
			"message": message,
			"details": details,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(response)
}

func (ws *webServer) NewClient(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Processing new client request")

	var req controller.NewClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode request: %v", err)
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	res, err := ws.app.ClientCtrl.NewClient(req)
	if err != nil {
		logger.Errorf("Failed to create client: %v", err)
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create client", nil)
		return
	}

	logger.WithFields(logrus.Fields{
		"username": req.Username,
	}).Info("Client created successfully")

	JSONSuccess(w, http.StatusCreated, "Client created successfully", res)
}

func (ws *webServer) Login(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Processing login request")

	var req controller.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Errorf("Failed to decode request: %v", err)
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	res, err := ws.app.ClientCtrl.Login(req)
	if err != nil {
		logger.Errorf("Failed to login: %v", err)
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid credentials", nil)
		return
	}

	logger.WithFields(logrus.Fields{
		"username": req.Username,
	}).Info("Login successful")

	JSONSuccess(w, http.StatusOK, "Login successful", res)
}
