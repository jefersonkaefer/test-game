package game

import (
	"context"
	"dicegame/auth"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Em produção, implemente uma verificação adequada
	},
}

type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type WalletRequest struct {
	ClientID string `json:"clientId"`
}

type PlayRequest struct {
	Amount  float64 `json:"amount"`
	BetType string  `json:"betType"`
}

type EndPlayRequest struct {
	ClientID string `json:"clientId"`
}

type Response struct {
	Type    string      `json:"type"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func (gm *GameManager) HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Obtém o token da URL
	vars := mux.Vars(r)
	token := vars["token"]
	if token == "" {
		http.Error(w, "Token não fornecido", http.StatusUnauthorized)
		return
	}

	// Valida o token
	claims, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "Token inválido", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Erro ao atualizar conexão:", err)
		return
	}
	defer conn.Close()

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Erro ao ler mensagem:", err)
			return
		}

		var response Response
		switch msg.Type {
		case "wallet":
			player := gm.GetPlayer(claims.UserID)
			response = Response{
				Type:    "wallet",
				Success: true,
				Data:    map[string]float64{"balance": player.Balance},
			}

		case "play":
			var req PlayRequest
			if err := json.Unmarshal(msg.Payload, &req); err != nil {
				response = Response{Type: "play", Success: false, Error: "Formato inválido"}
			} else {
				player := gm.GetPlayer(claims.UserID)
				game, err := player.PlaceBet(req.Amount, req.BetType)
				if err != nil {
					response = Response{Type: "play", Success: false, Error: err.Error()}
				} else {
					response = Response{
						Type:    "play",
						Success: true,
						Data:    game,
					}
				}
			}

		case "endPlay":
			player := gm.GetPlayer(claims.UserID)
			err := player.EndGame()
			if err != nil {
				response = Response{Type: "endPlay", Success: false, Error: err.Error()}
			} else {
				response = Response{
					Type:    "endPlay",
					Success: true,
					Data:    map[string]float64{"balance": player.Balance},
				}
			}

		default:
			response = Response{Type: msg.Type, Success: false, Error: "Tipo de mensagem desconhecido"}
		}

		if err := conn.WriteJSON(response); err != nil {
			log.Println("Erro ao enviar resposta:", err)
			return
		}
	}
} 