package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/dehwyy/dbfx/internal/fakesql"
	"github.com/stretchr/testify/require"
)

func TestApplyPool(t *testing.T) {
	tests := []struct {
		name     string
		opts     Opts
		wantOpen int
	}{
		{
			name: "single dsn limits applied",
			opts: Opts{
				ConnectionMaxOpen:     7,
				ConnectionMaxIdle:     3,
				ConnectionMaxLifetime: time.Minute,
				ConnectionIdleTime:    time.Second,
			},
			wantOpen: 7,
		},
		{
			name:     "zero keeps database/sql default",
			opts:     Opts{},
			wantOpen: 0,
		},
		{
			name: "only max open",
			opts: Opts{
				ConnectionMaxOpen: 30,
			},
			wantOpen: 30,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				db := fakesql.Open(
					t,
					"ok",
				)
				t.Cleanup(func() {
					require.NoError(
						t,
						db.Close(),
					)
				})

				applyPool(
					db,
					tt.opts,
				)

				require.Equal(
					t,
					tt.wantOpen,
					db.Stats().MaxOpenConnections,
				)
			},
		)
	}
}

func TestPing(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		timeout time.Duration
		wantErr error
	}{
		{
			name:    "ok",
			mode:    "ok",
			timeout: time.Second,
		},
		{
			name: "ok with default timeout",
			mode: "ok",
		},
		{
			name:    "driver error",
			mode:    "fail",
			timeout: time.Second,
			wantErr: fakesql.ErrPing,
		},
		{
			name:    "timeout",
			mode:    "hang",
			timeout: 50 * time.Millisecond,
			wantErr: context.DeadlineExceeded,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				db := fakesql.Open(
					t,
					tt.mode,
				)
				t.Cleanup(func() {
					require.NoError(
						t,
						db.Close(),
					)
				})

				err := Ping(
					db,
					tt.timeout,
				)

				if tt.wantErr == nil {
					require.NoError(
						t,
						err,
					)
					return
				}
				require.ErrorIs(
					t,
					err,
					tt.wantErr,
				)
			},
		)
	}
}

func TestNewRejectsEmptyDSN(t *testing.T) {
	_, err := New(Opts{})
	require.Error(
		t,
		err,
	)
}
