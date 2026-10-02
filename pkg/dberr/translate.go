package dberr

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

const (
	codeUniqueViolation     = "23505"
	codeForeignKeyViolation = "23503"
	codeNotNullViolation    = "23502"
	codeCheckViolation      = "23514"
	codeSerialization       = "40001"
	codeDeadlock            = "40P01"
)

func Translate(err error) error {
	if err == nil {
		return nil
	}

	if target := target(err); target != nil {
		return fmt.Errorf(
			"%w: %w",
			target,
			err,
		)
	}

	return err
}

func target(err error) error {
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		return ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return ErrConflict
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return ErrReference
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return nil
	}

	switch pgErr.Code {
	case codeUniqueViolation:
		return ErrConflict
	case codeForeignKeyViolation:
		return ErrReference
	case codeNotNullViolation:
		return ErrNotNull
	case codeCheckViolation:
		return ErrCheck
	case codeSerialization, codeDeadlock:
		return ErrRetryable
	}

	return nil
}
