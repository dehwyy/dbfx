package postgres

import (
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Opts struct {
	ConnectionStrings     []string
	ConnectionIdleTime    time.Duration // 10 minutes
	ConnectionMaxLifetime time.Duration // 30 minutes
	ConnectionMaxIdle     int           // 10
	ConnectionMaxOpen     int           // 30
}

func New(opts Opts) (*gorm.DB, error) {
	if len(opts.ConnectionStrings) == 0 {
		return nil, errors.New("connection strings slice is empty")
	}

	conn, err := gorm.Open(
		postgres.Open(
			opts.ConnectionStrings[0],
		),
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return conn, nil
}
