package errs

import "errors"

var (
	ErrUsernameExists = errors.New("username already exists")
	ErrNotFound       = errors.New("not found")
)
