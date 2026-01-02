package postgres

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Opts struct {
	// 1st is rw, others are read
	ConnectionStrings []string
}

func New(opts Opts) (*gorm.DB, error) {
	if len(opts.ConnectionStrings) == 0 {
		return nil, ErrConnectionStringIsEmpty
	}

	conn, err := gorm.Open(
		postgres.Open(
			opts.ConnectionStrings[0],
		),
	)
	if err != nil {
		return nil, err
	}

	// If only 1 connection, no read replicas
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

	if err := conn.Use(
		dbresolver.
			Register(dbresolver.Config{
				Replicas: replicas,
				Policy:   dbresolver.RandomPolicy{},
			}).
			SetConnMaxIdleTime(10 * time.Minute).
			SetConnMaxLifetime(30 * time.Minute).
			SetMaxIdleConns(10).
			SetMaxOpenConns(30),
	); err != nil {
		return nil, err
	}

	return conn, nil
}
