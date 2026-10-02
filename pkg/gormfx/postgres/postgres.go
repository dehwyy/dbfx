package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const DefaultPingTimeout = 10 * time.Second

type Opts struct {
	ConnectionStrings     []string
	ConnectionIdleTime    time.Duration
	ConnectionMaxLifetime time.Duration
	ConnectionMaxIdle     int
	ConnectionMaxOpen     int
	PingTimeout           time.Duration
	TranslateError        bool
}

func New(opts Opts) (*gorm.DB, error) {
	if len(opts.ConnectionStrings) == 0 {
		return nil, errors.New("connection strings slice is empty")
	}

	conn, err := gorm.Open(
		postgres.Open(
			opts.ConnectionStrings[0],
		),
		&gorm.Config{
			TranslateError:       opts.TranslateError,
			DisableAutomaticPing: true,
		},
	)
	if err != nil {
		return nil, err
	}

	sqlDB, err := conn.DB()
	if err != nil {
		return nil, err
	}

	applyPool(sqlDB, opts)

	if err := Ping(sqlDB, opts.PingTimeout); err != nil {
		return nil, errors.Join(err, sqlDB.Close())
	}

	if len(opts.ConnectionStrings) == 1 {
		return conn, nil
	}

	replicas := make([]gorm.Dialector, 0)
	for _, dsn := range opts.ConnectionStrings[1:] {
		if dsn == "" {
			continue
		}
		replicas = append(replicas, postgres.Open(dsn))
	}

	resolver := dbresolver.Register(
		dbresolver.Config{
			Replicas: replicas,
			Policy:   dbresolver.RandomPolicy{},
		},
	)

	if opts.ConnectionIdleTime > 0 {
		resolver.SetConnMaxIdleTime(opts.ConnectionIdleTime)
	}
	if opts.ConnectionMaxLifetime > 0 {
		resolver.SetConnMaxLifetime(opts.ConnectionMaxLifetime)
	}
	if opts.ConnectionMaxIdle > 0 {
		resolver.SetMaxIdleConns(opts.ConnectionMaxIdle)
	}
	if opts.ConnectionMaxOpen > 0 {
		resolver.SetMaxOpenConns(opts.ConnectionMaxOpen)
	}

	if err := conn.Use(resolver); err != nil {
		return nil, errors.Join(fmt.Errorf("register dbresolver: %w", err), sqlDB.Close())
	}

	return conn, nil
}

func applyPool(db *sql.DB, opts Opts) {
	if opts.ConnectionIdleTime > 0 {
		db.SetConnMaxIdleTime(opts.ConnectionIdleTime)
	}
	if opts.ConnectionMaxLifetime > 0 {
		db.SetConnMaxLifetime(opts.ConnectionMaxLifetime)
	}
	if opts.ConnectionMaxIdle > 0 {
		db.SetMaxIdleConns(opts.ConnectionMaxIdle)
	}
	if opts.ConnectionMaxOpen > 0 {
		db.SetMaxOpenConns(opts.ConnectionMaxOpen)
	}
}
