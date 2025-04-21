package errs

import "errors"

var (
	ErrUsernameExists       = errors.New("username already exists")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrNotFound             = errors.New("not found")
	ErrInsufficientBalance  = errors.New("insufficient balance")
	ErrPlayerAlreadyInMatch = errors.New("player already in match")
	ErrPlayerNotInMatch     = errors.New("player not in match")
)
