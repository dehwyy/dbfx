package migrate

import (
	"go.uber.org/fx"
	"gorm.io/gorm"
)

func ModuleConfig(cfg Config) fx.Option {
	return fx.Provide(
		func(db *gorm.DB) (*Migrator, error) {
			return fromGorm(
				db,
				cfg,
			)
		},
	)
}
