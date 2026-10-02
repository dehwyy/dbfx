package fakesql

import (
	"database/sql"
	"testing"
)

func Open(t *testing.T, mode string) *sql.DB {
	t.Helper()

	registerOnce.Do(func() {
		sql.Register(
			driverName,
			fakeDriver{},
		)
	})

	db, err := sql.Open(
		driverName,
		mode,
	)
	if err != nil {
		t.Fatalf(
			"open fake db: %v",
			err,
		)
	}

	return db
}
