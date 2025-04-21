package dto

type ClientLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateClientRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type CreateClientResponse struct {
	ID string `json:"id"`
}

type GetBalanceResponse struct {
	Balance float64 `json:"balance"`
}
