package postgres

import (
	"context"
	"fmt"
	"log/slog"
)

func (l *sqlLogger) emit(ctx context.Context, level slog.Level, msg string, data []any) {
	if len(data) > 0 {
		msg = fmt.Sprintf(
			msg,
			data...,
		)
	}

	attrs := make(
		[]slog.Attr,
		0,
		1,
	)
	if caller := callerFile(); caller != "" {
		attrs = append(
			attrs,
			slog.String(
				"caller",
				caller,
			),
		)
	}

	l.log.LogAttrs(
		ctx,
		level,
		msg,
		attrs...,
	)
}
