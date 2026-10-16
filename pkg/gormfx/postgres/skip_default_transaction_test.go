package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolveSkipDefaultTransaction(t *testing.T) {
	tests := []struct {
		name     string
		explicit bool
		env      string
		want     bool
		wantErr  error
	}{
		{
			name: "default is off",
			want: false,
		},
		{
			name:     "explicit field",
			explicit: true,
			want:     true,
		},
		{
			name: "env true",
			env:  "true",
			want: true,
		},
		{
			name: "env one with spaces",
			env:  " 1 ",
			want: true,
		},
		{
			name: "env false",
			env:  "false",
			want: false,
		},
		{
			name:     "field wins over env false",
			explicit: true,
			env:      "false",
			want:     true,
		},
		{
			name:    "broken env",
			env:     "maybe",
			wantErr: ErrInvalidEnv,
		},
		{
			name:     "broken env fails even with field",
			explicit: true,
			env:      "maybe",
			wantErr:  ErrInvalidEnv,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				got, err := resolveSkipDefaultTransaction(
					tt.explicit,
					envGetter(map[string]string{
						EnvSkipDefaultTransaction: tt.env,
					}),
				)

				if tt.wantErr != nil {
					require.ErrorIs(
						t,
						err,
						tt.wantErr,
					)
					return
				}
				require.NoError(
					t,
					err,
				)
				require.Equal(
					t,
					tt.want,
					got,
				)
			},
		)
	}
}

func TestNewSkipDefaultTransaction(t *testing.T) {
	tests := []struct {
		name    string
		opts    Opts
		env     string
		want    bool
		wantErr error
	}{
		{
			name: "off without field and env",
			want: false,
		},
		{
			name: "on by field",
			opts: Opts{
				SkipDefaultTransaction: true,
			},
			want: true,
		},
		{
			name: "on by env",
			env:  "true",
			want: true,
		},
		{
			name:    "broken env fails New",
			env:     "maybe",
			wantErr: ErrInvalidEnv,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				if tt.env != "" {
					t.Setenv(
						EnvSkipDefaultTransaction,
						tt.env,
					)
				}

				opts := tt.opts
				opts.ConnectionStrings = []string{startFakePG(t)}
				opts.PingTimeout = 5 * time.Second

				conn, err := New(opts)

				if tt.wantErr != nil {
					require.ErrorIs(
						t,
						err,
						tt.wantErr,
					)
					require.Nil(
						t,
						conn,
					)
					return
				}
				require.NoError(
					t,
					err,
				)
				sqlDB, err := conn.DB()
				require.NoError(
					t,
					err,
				)
				t.Cleanup(func() {
					require.NoError(
						t,
						sqlDB.Close(),
					)
				})

				require.Equal(
					t,
					tt.want,
					conn.Config.SkipDefaultTransaction,
				)
			},
		)
	}
}
