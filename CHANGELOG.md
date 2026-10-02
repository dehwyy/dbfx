# Changelog

## v0.2.0

### Fixed
- Pool limits (`ConnectionMaxOpen`, `ConnectionMaxIdle`, `ConnectionMaxLifetime`, `ConnectionIdleTime`) were ignored with a single DSN. They are now applied to the primary `sql.DB`.

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
- `go` directive raised to 1.25.7 (goose requirement). New dependencies: `github.com/pressly/goose/v3`, `go.uber.org/fx`.

### Not included
- Multi-DSN failover: `ConnectionStrings[1:]` stay read replicas through dbresolver.

## v0.1.0
- Initial `gormfx` / `postgres` constructor.
