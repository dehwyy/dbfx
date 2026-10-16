package postgres

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gormlogger "gorm.io/gorm/logger"
)

func envGetter(env map[string]string) func(string) string {
	return func(name string) string {
		return env[name]
	}
}

func TestResolveLogSettings(t *testing.T) {
	out := &bytes.Buffer{}

	tests := []struct {
		name       string
		opts       Opts
		env        map[string]string
		wantLevel  gormlogger.LogLevel
		wantSlow   time.Duration
		wantErr    error
		wantOutput bool
	}{
		{
			name:      "defaults",
			wantLevel: gormlogger.Warn,
			wantSlow:  200 * time.Millisecond,
		},
		{
			name: "env level and threshold",
			env: map[string]string{
				EnvLogLevel:         "info",
				EnvLogSlowThreshold: "1s",
			},
			wantLevel: gormlogger.Info,
			wantSlow:  time.Second,
		},
		{
			name: "env level is case and space tolerant",
			env: map[string]string{
				EnvLogLevel: "  SiLeNt ",
			},
			wantLevel: gormlogger.Silent,
			wantSlow:  200 * time.Millisecond,
		},
		{
			name: "env error level",
			env: map[string]string{
				EnvLogLevel: "error",
			},
			wantLevel: gormlogger.Error,
			wantSlow:  200 * time.Millisecond,
		},
		{
			name: "env zero threshold disables slow",
			env: map[string]string{
				EnvLogSlowThreshold: "0",
			},
			wantLevel: gormlogger.Warn,
			wantSlow:  0,
		},
		{
			name: "opts beat env",
			opts: Opts{
				LogLevel:         gormlogger.Error,
				LogSlowThreshold: 50 * time.Millisecond,
			},
			env: map[string]string{
				EnvLogLevel:         "info",
				EnvLogSlowThreshold: "1s",
			},
			wantLevel: gormlogger.Error,
			wantSlow:  50 * time.Millisecond,
		},
		{
			name: "negative opts threshold disables slow over env",
			opts: Opts{
				LogSlowThreshold: -1,
			},
			env: map[string]string{
				EnvLogSlowThreshold: "1s",
			},
			wantLevel: gormlogger.Warn,
			wantSlow:  0,
		},
		{
			name: "opts skip broken env",
			opts: Opts{
				LogLevel:         gormlogger.Warn,
				LogSlowThreshold: time.Second,
			},
			env: map[string]string{
				EnvLogLevel:         "loud",
				EnvLogSlowThreshold: "soon",
			},
			wantLevel: gormlogger.Warn,
			wantSlow:  time.Second,
		},
		{
			name: "custom output",
			opts: Opts{
				LogOutput: out,
			},
			wantLevel:  gormlogger.Warn,
			wantSlow:   200 * time.Millisecond,
			wantOutput: true,
		},
		{
			name: "broken env level",
			env: map[string]string{
				EnvLogLevel: "loud",
			},
			wantErr: ErrInvalidEnv,
		},
		{
			name: "broken env threshold",
			env: map[string]string{
				EnvLogSlowThreshold: "soon",
			},
			wantErr: ErrInvalidEnv,
		},
		{
			name: "negative env threshold",
			env: map[string]string{
				EnvLogSlowThreshold: "-1s",
			},
			wantErr: ErrInvalidEnv,
		},
		{
			name: "invalid opts level",
			opts: Opts{
				LogLevel: gormlogger.LogLevel(42),
			},
			wantErr: ErrInvalidOpts,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				settings, err := resolveLogSettings(
					tt.opts,
					envGetter(tt.env),
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
					tt.wantLevel,
					settings.level,
				)
				require.Equal(
					t,
					tt.wantSlow,
					settings.slow,
				)
				if tt.wantOutput {
					require.Same(
						t,
						out,
						settings.output,
					)
					return
				}
				require.Same(
					t,
					os.Stdout,
					settings.output,
				)
			},
		)
	}
}
