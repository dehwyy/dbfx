package postgres

import (
	"context"
	"log/slog"

	gormlogger "gorm.io/gorm/logger"
)

func (l *sqlLogger) Error(ctx context.Context, msg string, data ...any) {
	if l.level < gormlogger.Error {
		return
	}

	l.emit(
		ctx,
		slog.LevelError,
		msg,
		data,
	)
}
