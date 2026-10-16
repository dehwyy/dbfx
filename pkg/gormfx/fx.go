package gormfx

import (
	"errors"
	"time"

	"github.com/dehwyy/dbfx/pkg/gormfx/postgres"
	"gorm.io/gorm"
)

type PostgresOpts = postgres.Opts

type Opts struct {
	Postgres               *PostgresOpts
	PingTimeout            time.Duration
	TranslateError         bool
	SkipDefaultTransaction bool
}

func New(opts Opts) func() (*gorm.DB, error) {
	return func() (*gorm.DB, error) {
		switch {
		case opts.Postgres != nil:
			return postgres.New(
				mergePostgres(opts),
			)
		default:
			return nil, errors.New("no database provided")
		}
	}
}
