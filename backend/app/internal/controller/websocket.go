package controller

import (
	"encoding/json"
	"log"
)

const (
	ActionNewPlayer = "new_player"
	ActionNewGame   = "new_game"
	ActionNewMatch  = "new_match"
)

func (a *App) WebSocket(req Request) (res Response) {
	jsonBody, err := json.Marshal(req.Body)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	switch req.Action {
	case ActionNewPlayer:
		var r NewClientRequest
		if err = json.Unmarshal(jsonBody, &r); err != nil {
			log.Default().Printf("ERRROOOOOOR  ction :::Action :::%v", req)
			return
		}
		a.NewClient(r)
	}
	return Response{}
}
