package postgres

import (
	"log/slog"
	"strings"
	"time"
)

func replaceLogAttr(groups []string, attr slog.Attr) slog.Attr {
	if len(groups) > 0 {
		return attr
	}

	switch attr.Key {
	case slog.TimeKey:
		attr.Value = slog.StringValue(
			attr.Value.Time().Format(time.RFC3339),
		)
	case slog.LevelKey:
		if level, ok := attr.Value.Any().(slog.Level); ok {
			attr.Value = slog.StringValue(
				strings.ToLower(level.String()),
			)
		}
	case slog.MessageKey:
		attr.Key = "message"
	}

	return attr
}
