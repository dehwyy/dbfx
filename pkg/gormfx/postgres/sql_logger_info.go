package postgres

import (
	"context"
	"log/slog"

	gormlogger "gorm.io/gorm/logger"
)

func (l *sqlLogger) Info(ctx context.Context, msg string, data ...any) {
	if l.level < gormlogger.Info {
		return
	}

	l.emit(
		ctx,
		slog.LevelInfo,
		msg,
		data,
	)
}
