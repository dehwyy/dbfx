package postgres

import (
	"context"
	"log/slog"

	gormlogger "gorm.io/gorm/logger"
)

func (l *sqlLogger) Warn(ctx context.Context, msg string, data ...any) {
	if l.level < gormlogger.Warn {
		return
	}

	l.emit(
		ctx,
		slog.LevelWarn,
		msg,
		data,
	)
}
