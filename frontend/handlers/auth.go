package handlers

import (
	"context"
	"dicegame/auth"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

type AuthHandler struct {
	users map[string]*auth.User
}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		users: make(map[string]*auth.User),
	}
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Formato inválido", http.StatusBadRequest)
		return
	}

	// Verifica se o usuário já existe
	for _, user := range h.users {
		if user.Username == req.Username {
			http.Error(w, "Usuário já existe", http.StatusConflict)
			return
		}
	}

	// Hash da senha
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, "Erro ao criar usuário", http.StatusInternalServerError)
		return
	}

	// Cria novo usuário
	user := &auth.User{
		ID:       uuid.New().String(),
		Username: req.Username,
		Password: hashedPassword,
	}
	h.users[user.ID] = user

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Usuário criado com sucesso"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Formato inválido", http.StatusBadRequest)
		return
	}

	// Procura o usuário
	var user *auth.User
	for _, u := range h.users {
		if u.Username == req.Username {
			user = u
			break
		}
	}

	if user == nil || !auth.CheckPasswordHash(req.Password, user.Password) {
		http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
		return
	}

	// Gera o token JWT
	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Erro ao gerar token", http.StatusInternalServerError)
		return
	}

	response := AuthResponse{Token: token}
	json.NewEncoder(w).Encode(response)
}

func (h *AuthHandler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/register", h.Register).Methods("POST")
	r.HandleFunc("/login", h.Login).Methods("POST")
} 