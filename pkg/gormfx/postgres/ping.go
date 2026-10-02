package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func Ping(db *sql.DB, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = DefaultPingTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
