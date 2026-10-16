package postgres

import "time"

const (
	EnvLogLevel               = "DBFX_LOG_LEVEL"
	EnvLogSlowThreshold       = "DBFX_LOG_SLOW_THRESHOLD"
	EnvSkipDefaultTransaction = "DBFX_SKIP_DEFAULT_TRANSACTION"

	DefaultLogSlowThreshold = 200 * time.Millisecond
	MaxLoggedSQL            = 4096
)
