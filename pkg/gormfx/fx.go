package gormfx

import (
	"context"
	"fmt"

	"github.com/dehwyy/dbfx/pkg/gormfx/postgres"
	"github.com/samber/lo"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

type database string

const (
	DatabasePostgres database = "postgres"
)

type Opts struct {
	// Database kind
	Database database
	// Database connection string
	ConnectionStringsFromSecretsKey string
	// Would be sprintf'd for conn string
	ConnOptsFormat []any
}

type FxOpts struct {
	fx.In

	SecretsProvider SecretsProvider
}

func New(opts Opts) func(FxOpts) (*gorm.DB, error) {
	return func(fxOpts FxOpts) (*gorm.DB, error) {

		connectionStrings := fxOpts.SecretsProvider.MustGet(
			context.Background(),
			opts.ConnectionStringsFromSecretsKey,
		)
		connectionSlice, ok := connectionStrings.([]any)
		if !ok {
			return nil, ErrConnectionStringIsNotSlice
		}

		if len(connectionSlice) == 0 {
			return nil, ErrConnectionStringSliceIsEmpty
		}

		formattedConnectionStrings := lo.Map(
			connectionSlice,
			func(connStr any, _ int) string {
				return fmt.Sprintf(connStr.(string), opts.ConnOptsFormat...)
			},
		)

		switch opts.Database {
		case DatabasePostgres:
			return postgres.New(postgres.Opts{
				ConnectionStrings: formattedConnectionStrings,
			})
		default:
			return nil, ErrDatabaseNotFound
		}
	}
}
