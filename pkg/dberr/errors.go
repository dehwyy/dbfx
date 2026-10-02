package dberr

import "errors"

var (
	ErrNotFound  = errors.New("not found")
	ErrConflict  = errors.New("conflict")
	ErrReference = errors.New("reference violation")
	ErrRetryable = errors.New("retryable")
	ErrNotNull   = errors.New("not null violation")
	ErrCheck     = errors.New("check violation")
)
