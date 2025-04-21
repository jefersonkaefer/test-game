package database

import "encoding/json"

type PlayerData struct {
	ClientID string  `json:"client_id"`
	Balance  float64 `json:"balance"`
	InPlay   bool    `json:"in_play"`
}

func (p *PlayerData) MarshalBinary() ([]byte, error) {
	return json.Marshal(p)
}

func (p *PlayerData) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, p)
}
