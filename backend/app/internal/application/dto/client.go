package dto

type ClientLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ClientLoginResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

type CreateClientRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateClientResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ClientDTO struct {
	ID       string  `json:"id"`
	Username string  `json:"username"`
	Balance  float64 `json:"balance"`
}

type UpdateClientBalanceRequest struct {
	ClientID string  `json:"client_id"`
	Amount   float64 `json:"amount"`
}

type UpdateClientBalanceResponse struct {
	ID      string  `json:"id"`
	Balance float64 `json:"balance"`
}
