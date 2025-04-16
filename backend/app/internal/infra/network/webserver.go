package network

import (
	"encoding/json"
	"net/http"
	"time"

	"game/api/internal/controller"
	"game/api/internal/errs"
)

type webServer struct {
	app *controller.App
}

func NewWebServer(app *controller.App) *webServer {
	return &webServer{
		app: app,
	}
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
	w.WriteHeader(code)
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
	var req controller.NewClientRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}

	res, err := ws.app.NewClient(req)
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

func (ws *webServer) Login(w http.ResponseWriter, r *http.Request) {
	var req controller.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	res, err := ws.app.Login(req)
	if err != nil {
		if err == errs.ErrNotFound {
			JSONError(w, http.StatusNotFound,
				"CLIENT_NOT_FOUND",
				"This client was not found.",
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
	JSONSuccess(w, http.StatusOK,
		http.StatusText(http.StatusOK),
		res,
	)
}
