package fakesql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"sync"
	"testing"
)

var (
	ErrPing = errors.New("fake ping failed")

	registerOnce sync.Once
)

const driverName = "dbfx-fake"

type fakeDriver struct{}

func (fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{mode: name}, nil
}

type fakeConn struct {
	mode string
}

func (c *fakeConn) Prepare(string) (driver.Stmt, error) { return nil, driver.ErrSkip }
func (c *fakeConn) Close() error                        { return nil }
func (c *fakeConn) Begin() (driver.Tx, error)           { return nil, driver.ErrSkip }

func (c *fakeConn) Ping(ctx context.Context) error {
	switch c.mode {
	case "fail":
		return ErrPing
	case "hang":
		<-ctx.Done()
		return ctx.Err()
	}

	return nil
}

func Open(t *testing.T, mode string) *sql.DB {
	t.Helper()

	registerOnce.Do(func() {
		sql.Register(driverName, fakeDriver{})
	})

	db, err := sql.Open(driverName, mode)
	if err != nil {
		t.Fatalf("open fake db: %v", err)
	}

	return db
}
