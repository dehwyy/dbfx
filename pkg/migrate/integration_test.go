//go:build integration

package migrate_test

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/dehwyy/dbfx/pkg/gormfx"
	"github.com/dehwyy/dbfx/pkg/migrate"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openDB(t *testing.T, maxOpen int) *gorm.DB {
	t.Helper()

	dsn := os.Getenv("DBFX_TEST_DSN")
	if dsn == "" {
		t.Skip("DBFX_TEST_DSN is not set")
	}

	db, err := gormfx.New(gormfx.Opts{
		Postgres:       &gormfx.PostgresOpts{ConnectionStrings: []string{dsn}, ConnectionMaxOpen: maxOpen},
		TranslateError: true,
	})()
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, gormfx.Close(db)) })

	return db
}

func resetSchema(t *testing.T, db *gorm.DB, table string) {
	t.Helper()

	drop := func() {
		require.NoError(t, db.Exec("DROP TABLE IF EXISTS orders, "+table).Error)
	}
	drop()
	t.Cleanup(drop)
}

func newMigrator(t *testing.T, db *gorm.DB, table string, baseline []migrate.BaselineStep) *migrate.Migrator {
	t.Helper()

	sqlDB, err := db.DB()
	require.NoError(t, err)

	migrator, err := migrate.New(sqlDB, migrate.Config{
		FS:       os.DirFS("testdata/ok"),
		Table:    table,
		LockID:   73110001,
		Baseline: baseline,
	})
	require.NoError(t, err)

	return migrator
}

func TestPoolLimitAppliedWithSingleDSN(t *testing.T) {
	db := openDB(t, 5)

	sqlDB, err := db.DB()
	require.NoError(t, err)

	require.Equal(t, 5, sqlDB.Stats().MaxOpenConnections)
}

func TestUpFreshDatabase(t *testing.T) {
	db := openDB(t, 4)
	resetSchema(t, db, "dbfx_it_fresh")
	migrator := newMigrator(t, db, "dbfx_it_fresh", nil)

	applied, err := migrator.Up(context.Background())
	require.NoError(t, err)
	require.Len(t, applied, 2)

	again, err := migrator.Up(context.Background())
	require.NoError(t, err)
	require.Empty(t, again)

	var phase int
	require.NoError(t, db.Raw(
		"SELECT COUNT(*) FROM pg_attribute WHERE attrelid = 'orders'::regclass AND attname = 'balance_phase'",
	).Scan(&phase).Error)
	require.Equal(t, 1, phase)
}

func TestUpAdoptsLegacyDatabase(t *testing.T) {
	tests := []struct {
		name        string
		legacySQL   []string
		baseline    []migrate.BaselineStep
		wantApplied int
	}{
		{
			name:      "fully legacy applies nothing",
			legacySQL: []string{"CREATE TABLE orders (id bigserial PRIMARY KEY, balance_phase text)"},
			baseline: []migrate.BaselineStep{
				{Version: 1, Table: "orders"},
				{Version: 2, Table: "orders", Column: "balance_phase"},
			},
			wantApplied: 0,
		},
		{
			name:      "partial legacy applies the tail",
			legacySQL: []string{"CREATE TABLE orders (id bigserial PRIMARY KEY)"},
			baseline: []migrate.BaselineStep{
				{Version: 1, Table: "orders"},
				{Version: 2, Table: "orders", Column: "balance_phase"},
			},
			wantApplied: 1,
		},
		{
			name:        "fresh database ignores baseline",
			baseline:    []migrate.BaselineStep{{Version: 1, Table: "orders"}},
			wantApplied: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := openDB(t, 4)
			resetSchema(t, db, "dbfx_it_adopt")
			for _, statement := range tt.legacySQL {
				require.NoError(t, db.Exec(statement).Error)
			}
			migrator := newMigrator(t, db, "dbfx_it_adopt", tt.baseline)

			applied, err := migrator.Up(context.Background())
			require.NoError(t, err)
			require.Len(t, applied, tt.wantApplied)

			version, err := migrator.Version(context.Background())
			require.NoError(t, err)
			require.Equal(t, int64(2), version)
		})
	}
}

func TestConcurrentUpRunsEachMigrationOnce(t *testing.T) {
	db := openDB(t, 12)
	resetSchema(t, db, "dbfx_it_lock")

	const workers = 4
	var (
		wg    sync.WaitGroup
		mu    sync.Mutex
		total int
	)

	for range workers {
		migrator := newMigrator(t, db, "dbfx_it_lock", nil)
		wg.Add(1)
		go func() {
			defer wg.Done()

			applied, err := migrator.Up(context.Background())
			mu.Lock()
			defer mu.Unlock()
			require.NoError(t, err)
			total += len(applied)
		}()
	}
	wg.Wait()

	require.Equal(t, 2, total)
}
