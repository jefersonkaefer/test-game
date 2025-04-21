package entity

import "github.com/google/uuid"

type Wallet struct {
	ClientID uuid.UUID
	Balance  float64
}
