package dto

type MatchDTO struct {
	ID        string   `json:"id"`
	Players   []string `json:"players"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

type MatchPlayerDTO struct {
	ClientID string `json:"client_id"`
	Role     string `json:"role"`
}

type CreateMatchRequest struct {
	MinPlayers int    `json:"min_players"`
	MaxPlayers int    `json:"max_players"`
	GameMode   string `json:"game_mode"`
}

type CreateMatchResponse struct {
	ID        string   `json:"id"`
	Players   []string `json:"players"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

type GetMatchResponse struct {
	ID        string   `json:"id"`
	Players   []string `json:"players"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

type ListMatchesResponse struct {
	Matches []MatchDTO `json:"matches"`
}

type UpdateMatchRequest struct {
	Status string `json:"status"`
}

type UpdateMatchResponse struct {
	ID        string   `json:"id"`
	Players   []string `json:"players"`
	Status    string   `json:"status"`
	CreatedAt string   `json:"created_at"`
}

type AddPlayerRequest struct {
	MatchID string `json:"match_id"`
}
