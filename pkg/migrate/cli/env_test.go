package cli

import (
	"testing"
	"time"

	"github.com/dehwyy/dbfx/pkg/migrate"
	"github.com/stretchr/testify/require"
)

func TestParseBaseline(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []migrate.BaselineStep
		wantErr bool
	}{
		{name: "empty", raw: "", want: nil},
		{name: "tables", raw: "1:orders,2:providers", want: []migrate.BaselineStep{{Version: 1, Table: "orders"}, {Version: 2, Table: "providers"}}},
		{name: "column and spaces", raw: " 1:orders , 11:orders#balance_phase ", want: []migrate.BaselineStep{{Version: 1, Table: "orders"}, {Version: 11, Table: "orders", Column: "balance_phase"}}},
		{name: "no probe", raw: "1:orders,3:", want: []migrate.BaselineStep{{Version: 1, Table: "orders"}, {Version: 3}}},
		{name: "schema qualified", raw: "1:app.orders", want: []migrate.BaselineStep{{Version: 1, Table: "app.orders"}}},
		{name: "missing colon", raw: "1", wantErr: true},
		{name: "bad version", raw: "x:orders", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBaseline(tt.raw)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestLoadEnv(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		want    *Env
		wantErr bool
	}{
		{name: "missing dsn", env: map[string]string{EnvDir: "/m"}, wantErr: true},
		{name: "missing dir", env: map[string]string{EnvDSN: "postgres://x"}, wantErr: true},
		{name: "minimal", env: map[string]string{EnvDSN: "postgres://x", EnvDir: "/m"}, want: &Env{DSN: "postgres://x", Dir: "/m"}},
		{
			name: "full",
			env:  map[string]string{EnvDSN: "postgres://x", EnvDir: "/m", EnvTable: "v", EnvLockID: "42", EnvPingTimeout: "3s", EnvBaseline: "1:orders"},
			want: &Env{DSN: "postgres://x", Dir: "/m", Table: "v", LockID: 42, PingTimeout: 3 * time.Second, Baseline: []migrate.BaselineStep{{Version: 1, Table: "orders"}}},
		},
		{name: "bad lock id", env: map[string]string{EnvDSN: "postgres://x", EnvDir: "/m", EnvLockID: "abc"}, wantErr: true},
		{name: "bad ping timeout", env: map[string]string{EnvDSN: "postgres://x", EnvDir: "/m", EnvPingTimeout: "abc"}, wantErr: true},
		{name: "bad baseline", env: map[string]string{EnvDSN: "postgres://x", EnvDir: "/m", EnvBaseline: "zzz"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadEnv(func(key string) string { return tt.env[key] })

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
