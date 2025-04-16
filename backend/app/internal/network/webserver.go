package network

import (
	"encoding/json"
	"fmt"
	"net/http"

	"game/api/internal/controller"
)

type webServer struct {
	app *controller.App
}

func NewWebServer(app *controller.App) *webServer {
	return &webServer{
		app: app,
	}
}

func (ws *webServer) NewClient(w http.ResponseWriter, r *http.Request) {
	var req controller.NewClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	ws.app.NewClient(req)
}

func (ws *webServer) Login(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Olá, Mundo!")
}
