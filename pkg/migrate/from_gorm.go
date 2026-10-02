package migrate

import (
	"fmt"

	"gorm.io/gorm"
)

func fromGorm(db *gorm.DB, cfg Config) (*Migrator, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	return New(sqlDB, cfg)
}
