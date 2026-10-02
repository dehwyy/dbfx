package gormfx

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func ModuleOpts(opts Opts) fx.Option {
	return fx.Provide(
		func(lc fx.Lifecycle) (*gorm.DB, error) {
			return provide(
				lc,
				New(opts),
			)
		},
	)
}
