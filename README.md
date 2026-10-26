# dbfx

GORM + PostgreSQL helpers for fx services: constructor, fx module, error translation, goose migrations.

## gormfx

```go
gormfx.Module(func(cfg *Config) gormfx.Opts {
	return gormfx.Opts{
		Postgres: &gormfx.PostgresOpts{
			ConnectionStrings: []string{cfg.DSN},
			ConnectionMaxOpen: 20,
			ConnectionMaxIdle: 5,
		},
		TranslateError: true,
		PingTimeout:    5 * time.Second,
	}
})
```

`ConnectionStrings[0]` is the primary, the rest are read replicas (dbresolver). Pool limits apply to the primary and the replicas. The module closes the pool on `OnStop`. `gormfx.New(opts)()` stays available as before.

The SQL logger and the transaction option are configured through `DBFX_*` variables, see [gormfx environment](#gormfx-environment).

## dberr

```go
err := dberr.Translate(db.First(&row, id).Error)
if errors.Is(err, dberr.ErrNotFound) { ... }
```

`errors.Is` also matches the original gorm/pgx error.

## migrate

```go
//go:embed migrations/*.sql
var migrations embed.FS

fx.New(
	gormfx.Module(dbOpts),
	migrate.Module(func(cfg *Config) migrate.Config {
		return migrate.Config{FS: migrations, Dir: "migrations"}
	}),
	migrate.RunOnce(),
).Run()
```

Run it as a separate binary or initContainer, not in the API pod's `OnStart`. `Up` takes a session advisory lock, so concurrent runners are safe. A failed run exits with code 1.

### Adopting a database from a legacy sentinel runner

Name the files so that the legacy migrations come first (`00001_...`) and describe how to detect each of them:

```go
migrate.Config{
	FS:  migrations,
	Dir: "migrations",
	Baseline: []migrate.BaselineStep{
		{Version: 1, Table: "orders"},
		{Version: 2, Table: "providers"},
		{Version: 11, Table: "orders", Column: "balance_phase"},
	},
}
```

If the goose table has no applied versions, `Up` marks the longest prefix of steps whose probe object exists as applied, then runs the rest. A fresh database has no probe objects and runs everything. A step with an empty `Table` counts as applied when the previous step was. Take a backup before the first run on a real database.

### Stock binary

`cmd/dbfx-migrate` runs migrations from a directory:

| Variable | Meaning |
|---|---|
| `DBFX_DSN` | connection string (required) |
| `DBFX_MIGRATIONS_DIR` | directory with `.sql` files (required) |
| `DBFX_MIGRATIONS_TABLE` | version table, default `goose_db_version` |
| `DBFX_LOCK_ID` | advisory lock id, default goose's |
| `DBFX_BASELINE` | `1:orders,11:orders#balance_phase` |
| `DBFX_PING_TIMEOUT` | duration, default 10s |

## gormfx environment

`gormfx` installs its own GORM logger: one JSON line per record, `$N` placeholders without argument values, no `record not found`, no colors. The default level `warn` writes slow queries and errors only.

| Variable | Meaning | Opts field |
|---|---|---|
| `DBFX_LOG_LEVEL` | `silent`, `error`, `warn` (default), `info`; `info` writes every query, debugging only | `PostgresOpts.LogLevel` (`gorm.io/gorm/logger`) |
| `DBFX_LOG_SLOW_THRESHOLD` | duration, default `200ms`; `0` disables slow-query lines | `PostgresOpts.LogSlowThreshold` (negative disables) |
| `DBFX_SKIP_DEFAULT_TRANSACTION` | bool, default `false`; skips the implicit transaction around single writes | `Opts.SkipDefaultTransaction`, `PostgresOpts.SkipDefaultTransaction` |

A non-zero Opts field wins over the variable; a malformed variable makes `New` fail with `postgres.ErrInvalidEnv`. `PostgresOpts.LogOutput` replaces stdout.

```json
{"time":"2026-10-16T14:00:00Z","level":"warn","message":"slow sql","duration_ms":412.318,"rows":1,"caller":"/src/internal/repo/get.go:42","sql":"SELECT * FROM \"orders\" WHERE key = $1"}
```

`sql` is cut at 4096 bytes and `sql_len` then holds the original length. A logger that wraps this one must forward `ParamsFilter`, otherwise gorm hands the values back.

## Connection string from environment

```go
dsn, ok, err := postgres.ConnectionStringFromEnv("orders")
if err != nil {
	return err
}
if !ok {
	dsn = cfg.VaultDSN
}

gormfx.Opts{
	Postgres: &gormfx.PostgresOpts{
		ConnectionStrings: []string{dsn},
	},
}
```

`DB_PGSQL_HOST` switches the mode on; without it `ok` is `false` and the service keeps its old DSN (Vault). When it is set, the environment wins over `ConnectionStrings` that the service would otherwise take from Vault, so the service passes the returned DSN as `ConnectionStrings[0]`.

| Variable | Meaning |
|---|---|
| `DB_PGSQL_HOST` | host or IP; empty disables the mode |
| `DB_PGSQL_PORT` | default `5432` |
| `DB_PGSQL_DBNAME` | overrides the `dbName` argument |
| `DB_PGSQL_USER`, `DB_PGSQL_USERNAME` | either one; both set must be equal |
| `DB_PGSQL_PASSWORD` | required, taken as is |
| `DB_PGSQL_SSLMODE` | `disable`, `allow`, `prefer`, `require`, `verify-ca`, `verify-full` |
| `DB_PGSQL_SSLROOTCERT` | path to the CA bundle, passed as is |

Errors wrap `postgres.ErrInvalidEnv`, name the variable and never include its value.

## Tests

`go test ./...` needs no database. `go test -tags=integration ./...` with `DBFX_TEST_DSN` pointing to an empty throwaway PostgreSQL database also runs the migration and translation tests against it.
