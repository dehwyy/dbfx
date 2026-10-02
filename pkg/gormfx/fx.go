package gormfx

import (
	"errors"
	"time"

	"github.com/dehwyy/dbfx/pkg/gormfx/postgres"
	"gorm.io/gorm"
)

type PostgresOpts = postgres.Opts

type Opts struct {
	Postgres       *PostgresOpts
	PingTimeout    time.Duration
	TranslateError bool
}

func New(opts Opts) func() (*gorm.DB, error) {
	return func() (*gorm.DB, error) {
		switch {
		case opts.Postgres != nil:
			pg := *opts.Postgres
			if pg.PingTimeout == 0 {
				pg.PingTimeout = opts.PingTimeout
			}
			pg.TranslateError = pg.TranslateError || opts.TranslateError

			return postgres.New(pg)
		default:
			return nil, errors.New("no database provided")
		}
	}
}
