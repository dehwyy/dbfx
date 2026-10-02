package gormfx

import (
	"context"

	"go.uber.org/fx"
	"gorm.io/gorm"
)

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
