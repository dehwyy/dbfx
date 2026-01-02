package postgres

import "errors"

var (
	ErrConnectionStringIsEmpty = errors.New("connection string is empty")
)
