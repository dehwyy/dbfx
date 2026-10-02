package gormfx

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func Module[C any](opts func(*C) Opts) fx.Option {
	return fx.Provide(
		func(lc fx.Lifecycle, cfg *C) (*gorm.DB, error) {
			return provide(lc, New(opts(cfg)))
		},
	)
}
