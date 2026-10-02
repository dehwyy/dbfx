package gormfx

import (
	"fmt"

	"gorm.io/gorm"
)

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf(
			"get sql db: %w",
			err,
		)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf(
			"close sql db: %w",
			err,
		)
	}

	return nil
}
