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

## Tests

`go test ./...` needs no database. `go test -tags=integration ./...` with `DBFX_TEST_DSN` pointing to an empty throwaway PostgreSQL database also runs the migration and translation tests against it.
