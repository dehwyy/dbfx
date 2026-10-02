package cli

import (
	"os"

	"github.com/dehwyy/dbfx/pkg/gormfx"
	"github.com/dehwyy/dbfx/pkg/migrate"
	"go.uber.org/fx"
)

const poolSize = 2

var Module = fx.Options(
	fx.Provide(LoadEnv),
	gormfx.Module(
		func(env *Env) gormfx.Opts {
			return gormfx.Opts{
				Postgres: &gormfx.PostgresOpts{
					ConnectionStrings: []string{env.DSN},
					ConnectionMaxOpen: poolSize,
				},
				PingTimeout: env.PingTimeout,
			}
		},
	),
	migrate.Module(
		func(env *Env) migrate.Config {
			return migrate.Config{
				FS:       os.DirFS(env.Dir),
				Table:    env.Table,
				LockID:   env.LockID,
				Baseline: env.Baseline,
			}
		},
	),
	migrate.RunOnce(),
)
