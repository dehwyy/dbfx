package postgres

import (
	"log/slog"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	_ gormlogger.Interface = (*sqlLogger)(nil)
	_ gorm.ParamsFilter    = (*sqlLogger)(nil)
)

type sqlLogger struct {
	log   *slog.Logger
	level gormlogger.LogLevel
	slow  time.Duration
}

func newSQLLogger(settings logSettings) *sqlLogger {
	return &sqlLogger{
		log: slog.New(
			newLogHandler(settings.output),
		),
		level: settings.level,
		slow:  settings.slow,
	}
}
