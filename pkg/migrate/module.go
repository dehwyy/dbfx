package migrate

import (
	"fmt"

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

func ModuleConfig(cfg Config) fx.Option {
	return fx.Provide(
		func(db *gorm.DB) (*Migrator, error) {
			return fromGorm(db, cfg)
		},
	)
}

func fromGorm(db *gorm.DB, cfg Config) (*Migrator, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	return New(sqlDB, cfg)
}
