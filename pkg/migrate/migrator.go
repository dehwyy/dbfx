package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"
	"time"

	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

type Result struct {
	Version  int64
	Path     string
	Duration time.Duration
}

type Migrator struct {
	db       *sql.DB
	provider *goose.Provider
	cfg      Config
	lockID   int64
}

func New(db *sql.DB, cfg Config) (*Migrator, error) {
	cfg, err := cfg.validate()
	if err != nil {
		return nil, err
	}

	fsys := cfg.FS
	if cfg.Dir != "" {
		fsys, err = fs.Sub(cfg.FS, cfg.Dir)
		if err != nil {
			return nil, fmt.Errorf("sub fs %q: %w", cfg.Dir, err)
		}
	}

	lockID := cfg.LockID
	if lockID == 0 {
		lockID = lock.DefaultLockID
	}

	locker, err := lock.NewPostgresSessionLocker(lock.WithLockID(lockID))
	if err != nil {
		return nil, fmt.Errorf("create session locker: %w", err)
	}

	options := []goose.ProviderOption{
		goose.WithTableName(cfg.Table),
		goose.WithSessionLocker(locker),
	}
	if len(cfg.GoMigrations) > 0 {
		options = append(options, goose.WithGoMigrations(cfg.GoMigrations...))
	}
	if cfg.Logger != nil {
		options = append(options, goose.WithSlog(cfg.Logger))
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys, options...)
	if err != nil {
		return nil, fmt.Errorf("create goose provider: %w", err)
	}

	return &Migrator{
		db:       db,
		provider: provider,
		cfg:      cfg,
		lockID:   lockID,
	}, nil
}
