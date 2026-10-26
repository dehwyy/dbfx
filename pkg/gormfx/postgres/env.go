package postgres

import "time"

const (
	EnvLogLevel               = "DBFX_LOG_LEVEL"
	EnvLogSlowThreshold       = "DBFX_LOG_SLOW_THRESHOLD"
	EnvSkipDefaultTransaction = "DBFX_SKIP_DEFAULT_TRANSACTION"

	EnvPgsqlHost        = "DB_PGSQL_HOST"
	EnvPgsqlPort        = "DB_PGSQL_PORT"
	EnvPgsqlDBName      = "DB_PGSQL_DBNAME"
	EnvPgsqlUser        = "DB_PGSQL_USER"
	EnvPgsqlUsername    = "DB_PGSQL_USERNAME"
	EnvPgsqlPassword    = "DB_PGSQL_PASSWORD"
	EnvPgsqlSSLMode     = "DB_PGSQL_SSLMODE"
	EnvPgsqlSSLRootCert = "DB_PGSQL_SSLROOTCERT"

	DefaultPgsqlPort = 5432

	DefaultLogSlowThreshold = 200 * time.Millisecond
	MaxLoggedSQL            = 4096
)
