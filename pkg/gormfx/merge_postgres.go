package gormfx

import "github.com/dehwyy/dbfx/pkg/gormfx/postgres"

func mergePostgres(opts Opts) postgres.Opts {
	pg := *opts.Postgres
	if pg.PingTimeout == 0 {
		pg.PingTimeout = opts.PingTimeout
	}
	pg.TranslateError = pg.TranslateError || opts.TranslateError
	pg.SkipDefaultTransaction = pg.SkipDefaultTransaction || opts.SkipDefaultTransaction

	return pg
}
