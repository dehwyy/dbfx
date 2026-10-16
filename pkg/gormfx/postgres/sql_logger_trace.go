package postgres

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

func (l *sqlLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	failed := err != nil && !errors.Is(
		err,
		gormlogger.ErrRecordNotFound,
	)
	slow := l.slow > 0 && elapsed > l.slow

	var (
		level   slog.Level
		message string
	)

	switch {
	case failed && l.level >= gormlogger.Error:
		level, message = slog.LevelError, "sql error"
	case slow && l.level >= gormlogger.Warn:
		level, message = slog.LevelWarn, "slow sql"
	case l.level >= gormlogger.Info:
		level, message = slog.LevelInfo, "sql"
	default:
		return
	}

	sql, rows := fc()
	sql = normalizePlaceholders(sql)

	attrs := make(
		[]slog.Attr,
		0,
		7,
	)
	attrs = append(
		attrs,
		slog.Float64(
			"duration_ms",
			math.Round(elapsed.Seconds()*1e6)/1e3,
		),
	)
	if rows >= 0 {
		attrs = append(
			attrs,
			slog.Int64(
				"rows",
				rows,
			),
		)
	}
	if caller := callerFile(); caller != "" {
		attrs = append(
			attrs,
			slog.String(
				"caller",
				caller,
			),
		)
	}
	if len(sql) > MaxLoggedSQL {
		attrs = append(
			attrs,
			slog.Int(
				"sql_len",
				len(sql),
			),
		)
	}
	attrs = append(
		attrs,
		slog.String(
			"sql",
			truncateSQL(
				sql,
				MaxLoggedSQL,
			),
		),
	)
	if failed {
		attrs = append(
			attrs,
			slog.String(
				"error",
				err.Error(),
			),
		)
	}

	l.log.LogAttrs(
		ctx,
		level,
		message,
		attrs...,
	)
}
