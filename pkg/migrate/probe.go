package migrate

import (
	"context"
	"database/sql"
	"fmt"
)

const (
	tableExistsQuery  = "SELECT to_regclass($1) IS NOT NULL"
	columnExistsQuery = "SELECT EXISTS (SELECT 1 FROM pg_attribute WHERE attrelid = to_regclass($1) AND attname = $2 AND attnum > 0 AND NOT attisdropped)"
)

func probe(ctx context.Context, conn *sql.Conn, step BaselineStep) (bool, error) {
	var exists bool

	var err error
	if step.Column == "" {
		err = conn.QueryRowContext(
			ctx,
			tableExistsQuery,
			step.Table,
		).Scan(&exists)
	} else {
		err = conn.QueryRowContext(
			ctx,
			columnExistsQuery,
			step.Table,
			step.Column,
		).Scan(&exists)
	}
	if err != nil {
		return false, fmt.Errorf(
			"probe version %d: %w",
			step.Version,
			err,
		)
	}

	return exists, nil
}
