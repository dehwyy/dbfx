//go:build integration

package postgres

import (
	"os"
	"testing"
	"time"

	"github.com/dehwyy/dbfx/pkg/dberr"
	"github.com/stretchr/testify/require"
)

func dsn(t *testing.T) string {
	t.Helper()

	value := os.Getenv("DBFX_TEST_DSN")
	if value == "" {
		t.Skip("DBFX_TEST_DSN is not set")
	}

	return value
}

func TestNewTranslatesErrors(t *testing.T) {
	db, err := New(Opts{ConnectionStrings: []string{dsn(t)}, TranslateError: true, PingTimeout: 3 * time.Second, ConnectionMaxOpen: 1})
	require.NoError(t, err)

	require.NoError(t, db.Exec("CREATE TEMP TABLE dbfx_it_unique (id int PRIMARY KEY)").Error)

	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{name: "unique violation", run: func() error {
			if err := db.Exec("INSERT INTO dbfx_it_unique VALUES (1)").Error; err != nil {
				return err
			}
			return db.Exec("INSERT INTO dbfx_it_unique VALUES (1)").Error
		}, want: dberr.ErrConflict},
		{name: "record not found", run: func() error {
			var row struct{ ID int }
			return db.Table("dbfx_it_unique").Where("id = ?", 999).First(&row).Error
		}, want: dberr.ErrNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.ErrorIs(t, dberr.Translate(tt.run()), tt.want)
		})
	}
}

func TestNewFailsFastOnUnreachableHost(t *testing.T) {
	_, err := New(Opts{ConnectionStrings: []string{"postgres://u:p@127.0.0.1:1/db?sslmode=disable"}, PingTimeout: time.Second})

	require.Error(t, err)
}
