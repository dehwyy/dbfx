package gormfx

import (
	"errors"

	"github.com/dehwyy/dbfx/pkg/gormfx/postgres"
	"gorm.io/gorm"
)

type PostgresOpts = postgres.Opts

type Opts struct {
	Postgres *PostgresOpts
}

func New(opts Opts) func() (*gorm.DB, error) {
	return func() (*gorm.DB, error) {
		switch {
		case opts.Postgres != nil:
			return postgres.New(*opts.Postgres)
		default:
			return nil, errors.New("no database provided")
		}
	}
}
