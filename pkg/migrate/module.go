package migrate

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func Module[C any](cfg func(*C) Config) fx.Option {
	return fx.Provide(
		func(db *gorm.DB, appCfg *C) (*Migrator, error) {
			return fromGorm(db, cfg(appCfg))
		},
	)
}
