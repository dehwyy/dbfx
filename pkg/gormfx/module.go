package gormfx

import (
	"context"

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

func ModuleOpts(opts Opts) fx.Option {
	return fx.Provide(
		func(lc fx.Lifecycle) (*gorm.DB, error) {
			return provide(lc, New(opts))
		},
	)
}

func provide(lc fx.Lifecycle, open func() (*gorm.DB, error)) (*gorm.DB, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}

	lc.Append(
		fx.Hook{
			OnStop: func(context.Context) error {
				return Close(db)
			},
		},
	)

	return db, nil
}
