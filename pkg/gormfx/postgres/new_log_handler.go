package postgres

import (
	"io"
	"log/slog"
)

func newLogHandler(output io.Writer) slog.Handler {
	return slog.NewJSONHandler(
		output,
		&slog.HandlerOptions{
			Level:       slog.LevelDebug,
			ReplaceAttr: replaceLogAttr,
		},
	)
}
