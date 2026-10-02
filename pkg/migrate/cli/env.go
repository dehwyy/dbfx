package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/dehwyy/dbfx/pkg/migrate"
)

const (
	EnvDSN         = "DBFX_DSN"
	EnvDir         = "DBFX_MIGRATIONS_DIR"
	EnvTable       = "DBFX_MIGRATIONS_TABLE"
	EnvLockID      = "DBFX_LOCK_ID"
	EnvBaseline    = "DBFX_BASELINE"
	EnvPingTimeout = "DBFX_PING_TIMEOUT"
)

var ErrMissingEnv = errors.New("required environment variable is empty")

type Env struct {
	DSN         string
	Dir         string
	Table       string
	LockID      int64
	Baseline    []migrate.BaselineStep
	PingTimeout time.Duration
}

func LoadEnv() (*Env, error) {
	return loadEnv(os.Getenv)
}

func loadEnv(get func(string) string) (*Env, error) {
	env := &Env{
		DSN:   get(EnvDSN),
		Dir:   get(EnvDir),
		Table: get(EnvTable),
	}

	for name, value := range map[string]string{EnvDSN: env.DSN, EnvDir: env.Dir} {
		if value == "" {
			return nil, fmt.Errorf("%w: %s", ErrMissingEnv, name)
		}
	}

	if raw := get(EnvLockID); raw != "" {
		lockID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", EnvLockID, err)
		}
		env.LockID = lockID
	}

	if raw := get(EnvPingTimeout); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", EnvPingTimeout, err)
		}
		env.PingTimeout = timeout
	}

	baseline, err := ParseBaseline(get(EnvBaseline))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", EnvBaseline, err)
	}
	env.Baseline = baseline

	return env, nil
}

func ParseBaseline(raw string) ([]migrate.BaselineStep, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	steps := make([]migrate.BaselineStep, 0, len(parts))

	for _, part := range parts {
		versionRaw, probe, found := strings.Cut(strings.TrimSpace(part), ":")
		if !found {
			return nil, fmt.Errorf("step %q: expected version:table[#column]", part)
		}

		version, err := strconv.ParseInt(strings.TrimSpace(versionRaw), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("step %q: %w", part, err)
		}

		step := migrate.BaselineStep{Version: version}
		probe = strings.TrimSpace(probe)
		if probe != "" {
			step.Table, step.Column = splitProbe(probe)
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func splitProbe(probe string) (string, string) {
	table, column, found := strings.Cut(probe, "#")
	if found {
		return table, column
	}

	return probe, ""
}
