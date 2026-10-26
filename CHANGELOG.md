# Changelog

## v0.4.0

### Added
- `postgres.ConnectionStringFromEnv(dbName)` returns `(dsn, ok, err)`: a `postgres://` URL built from the `DB_PGSQL_*` variables (`HOST`, `PORT`, `DBNAME`, `USER`/`USERNAME`, `PASSWORD`, `SSLMODE`, `SSLROOTCERT`). An empty `DB_PGSQL_HOST` returns `ok=false` and nothing else is read. The user, password and database name are escaped by `net/url`, so any password character is safe.
- Defaults and rules: port `5432` (`1..65535`), database from `DB_PGSQL_DBNAME` or the `dbName` argument, user from `DB_PGSQL_USER` or `DB_PGSQL_USERNAME` (both set and different is an error), the password is required and never trimmed, `DB_PGSQL_SSLMODE` is one of `disable|allow|prefer|require|verify-ca|verify-full`.
- A malformed variable returns an error wrapping `postgres.ErrInvalidEnv`. The message names the variable and never contains a value.
- Constants `postgres.EnvPgsql*` and `postgres.DefaultPgsqlPort`.

## v0.3.0

### Added
- `postgres.Opts.LogLevel`, `LogSlowThreshold`, `LogOutput`: control of the SQL logger. Zero values fall back to the environment, then to the defaults. A negative `LogSlowThreshold` disables slow-query lines. `LogOutput` replaces stdout (tests).
- `DBFX_LOG_LEVEL` (`silent|error|warn|info`, default `warn`) and `DBFX_LOG_SLOW_THRESHOLD` (duration, default `200ms`, `0` disables slow-query lines). A malformed value makes `New` return `postgres.ErrInvalidEnv`.
- `postgres.Opts.SkipDefaultTransaction` and `gormfx.Opts.SkipDefaultTransaction` (default `false`) feed `gorm.Config.SkipDefaultTransaction`. `DBFX_SKIP_DEFAULT_TRANSACTION` (bool, default `false`) enables it without a code change; the option or the env turns it on. Explicit transactions are not affected. Enable it per service only after checking `Create`/`Save` with associations and hooks.
- `postgres.ErrInvalidEnv`, `postgres.ErrInvalidOpts`.

### Changed
- The default GORM logger is replaced by a dbfx logger on every `*gorm.DB` built by `gormfx`/`postgres.New`:
  - one JSON line per record on stdout (`time`, `level`, `message`, `duration_ms`, `rows`, `caller`, `sql`, `error`), no colors, no second line;
  - SQL keeps `$N` placeholders, argument values are never written (`ParamsFilter` drops them), so idempotency keys, API key ids, amounts and requisites stay out of the logs;
  - `gorm.ErrRecordNotFound` is not logged at the default level;
  - default level is `warn`: slow queries and errors only; a query that is neither is not rendered at all (`fc` is not called);
  - `sql` is cut at 4096 bytes on a rune boundary, `sql_len` carries the original length when it was cut.
- Session loggers (`Session(&gorm.Session{Logger: ...})`), `LogMode` and wrappers that forward `ParamsFilter` keep working. A wrapper that embeds `logger.Interface` but does not forward `ParamsFilter` makes gorm pass the values back to the logger; forward it.
- Log messages are `slow sql`, `sql error` and `sql`; the old `SLOW SQL >= ...` text and the `[rows:N]` suffix are gone.

### Not included
- Pool defaults (`MaxIdle`), acquire timeout, trace ids in log lines.
- The go directive and the gorm version are unchanged.

## v0.2.0

### Fixed
- Pool limits (`ConnectionMaxOpen`, `ConnectionMaxIdle`, `ConnectionMaxLifetime`, `ConnectionIdleTime`) were ignored with a single DSN. They are now applied to the primary `sql.DB`. Services that do not set these options keep the `database/sql` defaults (unlimited), so rollout changes nothing until limits are added to their configs.

### Added
- `gormfx.Module[C](func(*C) Opts)` and `gormfx.ModuleOpts(Opts)`: fx providers of `*gorm.DB` that close the pool on `OnStop`.
- `gormfx.Close(*gorm.DB)`.
- `Opts.PingTimeout` (default 10s): the primary is pinged with a bounded timeout at startup.
- `Opts.TranslateError`: enables gorm error translation (`gorm.ErrDuplicatedKey`, ...).
- `dberr`: `Translate(err)` maps gorm and pgx errors to `ErrNotFound`, `ErrConflict`, `ErrReference`, `ErrNotNull`, `ErrCheck`, `ErrRetryable`. The original error stays in the chain.
- `migrate`: goose-backed migrations (`embed.FS` or any `fs.FS`, session advisory lock, custom version table, Go migrations), `BaselineStep` adoption of databases migrated by legacy sentinel runners, `Module`/`ModuleConfig`/`RunOnce` for fx.
- `migrate/cli` and `cmd/dbfx-migrate`: env-driven one-shot migration binary.

### Changed
- The automatic gorm ping (unbounded) is replaced by an explicit ping with timeout. An unreachable host now fails after `PingTimeout`.
- `go` directive is 1.25.0 (pgx v5.10 floor); goose is pinned to v3.27.0 (go 1.25.0) because v3.28 requires go 1.26.0. Consumers building on go >= 1.25.0 images are unaffected. New dependencies: `github.com/pressly/goose/v3`, `go.uber.org/fx`.

### Not included
- Multi-DSN failover: `ConnectionStrings[1:]` stay read replicas through dbresolver.

## v0.1.0
- Initial `gormfx` / `postgres` constructor.
