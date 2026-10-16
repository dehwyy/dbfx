package postgres

import "errors"

var (
	ErrInvalidEnv  = errors.New("invalid environment variable")
	ErrInvalidOpts = errors.New("invalid options")
)
