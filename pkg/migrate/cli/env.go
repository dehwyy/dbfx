package cli

import (
	"errors"
	"fmt"
	"os"
	"strconv"
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
			return nil, fmt.Errorf(
				"%w: %s",
				ErrMissingEnv,
				name,
			)
		}
	}

	if raw := get(EnvLockID); raw != "" {
		lockID, err := strconv.ParseInt(
			raw,
			10,
			64,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"parse %s: %w",
				EnvLockID,
				err,
			)
		}
		env.LockID = lockID
	}

	if raw := get(EnvPingTimeout); raw != "" {
		timeout, err := time.ParseDuration(raw)
		if err != nil {
			return nil, fmt.Errorf(
				"parse %s: %w",
				EnvPingTimeout,
				err,
			)
		}
		env.PingTimeout = timeout
	}

	baseline, err := ParseBaseline(get(EnvBaseline))
	if err != nil {
		return nil, fmt.Errorf(
			"parse %s: %w",
			EnvBaseline,
			err,
		)
	}
	env.Baseline = baseline

	return env, nil
}
