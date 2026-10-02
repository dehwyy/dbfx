package migrate

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/pressly/goose/v3/lock"
)

func (m *Migrator) Adopt(ctx context.Context) ([]int64, error) {
	if len(m.cfg.Baseline) == 0 {
		return nil, nil
	}

	if _, err := m.provider.GetDBVersion(ctx); err != nil {
		return nil, fmt.Errorf("ensure version table: %w", err)
	}

	locker, err := lock.NewPostgresSessionLocker(lock.WithLockID(m.lockID))
	if err != nil {
		return nil, fmt.Errorf("create session locker: %w", err)
	}

	conn, err := m.db.Conn(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire connection: %w", err)
	}

	if err := locker.SessionLock(ctx, conn); err != nil {
		return nil, errors.Join(fmt.Errorf("lock: %w", err), conn.Close())
	}

	adopted, adoptErr := m.adoptLocked(ctx, conn)

	unlockErr := locker.SessionUnlock(context.WithoutCancel(ctx), conn)

	return adopted, errors.Join(adoptErr, unlockErr, conn.Close())
}

func (m *Migrator) adoptLocked(ctx context.Context, conn *sql.Conn) ([]int64, error) {
	var managed int
	err := conn.QueryRowContext(
		ctx,
		fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE version_id > 0 AND is_applied", m.cfg.Table),
	).Scan(&managed)
	if err != nil {
		return nil, fmt.Errorf("count applied: %w", err)
	}
	if managed > 0 {
		return nil, nil
	}

	known := make(map[int64]struct{})
	for _, version := range m.Versions() {
		known[version] = struct{}{}
	}

	versions, err := planAdopt(
		m.cfg.Baseline,
		known,
		func(step BaselineStep) (bool, error) {
			return probe(ctx, conn, step)
		},
	)
	if err != nil {
		return nil, err
	}
	if len(versions) == 0 {
		return nil, nil
	}

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin: %w", err)
	}

	for _, version := range versions {
		_, err := tx.ExecContext(
			ctx,
			fmt.Sprintf("INSERT INTO %s (version_id, is_applied) VALUES ($1, TRUE)", m.cfg.Table),
			version,
		)
		if err != nil {
			return nil, errors.Join(fmt.Errorf("mark version %d: %w", version, err), tx.Rollback())
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return versions, nil
}
