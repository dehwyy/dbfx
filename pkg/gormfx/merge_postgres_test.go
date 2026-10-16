package gormfx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMergePostgres(t *testing.T) {
	tests := []struct {
		name string
		opts Opts
		want PostgresOpts
	}{
		{
			name: "top level fills unset fields",
			opts: Opts{
				Postgres: &PostgresOpts{
					ConnectionStrings: []string{"dsn"},
				},
				PingTimeout:            3 * time.Second,
				TranslateError:         true,
				SkipDefaultTransaction: true,
			},
			want: PostgresOpts{
				ConnectionStrings:      []string{"dsn"},
				PingTimeout:            3 * time.Second,
				TranslateError:         true,
				SkipDefaultTransaction: true,
			},
		},
		{
			name: "postgres level wins for ping timeout and keeps flags",
			opts: Opts{
				Postgres: &PostgresOpts{
					ConnectionStrings:      []string{"dsn"},
					PingTimeout:            time.Second,
					TranslateError:         true,
					SkipDefaultTransaction: true,
				},
				PingTimeout: 3 * time.Second,
			},
			want: PostgresOpts{
				ConnectionStrings:      []string{"dsn"},
				PingTimeout:            time.Second,
				TranslateError:         true,
				SkipDefaultTransaction: true,
			},
		},
		{
			name: "everything off stays off",
			opts: Opts{
				Postgres: &PostgresOpts{
					ConnectionStrings: []string{"dsn"},
				},
			},
			want: PostgresOpts{
				ConnectionStrings: []string{"dsn"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				require.Equal(
					t,
					tt.want,
					mergePostgres(tt.opts),
				)
			},
		)
	}
}

func TestMergePostgresDoesNotMutateInput(t *testing.T) {
	pg := &PostgresOpts{
		ConnectionStrings: []string{"dsn"},
	}

	merged := mergePostgres(Opts{
		Postgres:               pg,
		SkipDefaultTransaction: true,
	})

	require.True(
		t,
		merged.SkipDefaultTransaction,
	)
	require.False(
		t,
		pg.SkipDefaultTransaction,
	)
}
