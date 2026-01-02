package gormfx

import "errors"

var (
	ErrDatabaseNotFound             = errors.New("provided database not found")
	ErrConnectionStringIsNotSlice   = errors.New("connection string is not a slice")
	ErrConnectionStringSliceIsEmpty = errors.New("connection string slice is empty")
)
