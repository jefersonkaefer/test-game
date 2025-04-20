package dto

type BetDTO struct {
	ID        string  `json:"id"`
	MatchID   string  `json:"match_id"`
	ClientID  string  `json:"client_id"`
	Amount    float64 `json:"amount"`
	Parity    string  `json:"parity"`
	CreatedAt string  `json:"created_at"`
}

type CreateBetRequest struct {
	MatchID string  `json:"match_id"`
	Amount  float64 `json:"amount"`
	Parity  string  `json:"parity"`
}

type GetBetResponse struct {
	ID        string  `json:"id"`
	MatchID   string  `json:"match_id"`
	ClientID  string  `json:"client_id"`
	Amount    float64 `json:"amount"`
	Parity    string  `json:"parity"`
	CreatedAt string  `json:"created_at"`
}

type ListBetsResponse struct {
	Bets []BetDTO `json:"bets"`
}
