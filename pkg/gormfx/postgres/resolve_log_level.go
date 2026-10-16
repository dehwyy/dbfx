package postgres

import (
	"fmt"
	"strings"

	gormlogger "gorm.io/gorm/logger"
)

func resolveLogLevel(explicit gormlogger.LogLevel, get func(string) string) (gormlogger.LogLevel, error) {
	if explicit != 0 {
		if explicit < gormlogger.Silent || explicit > gormlogger.Info {
			return 0, fmt.Errorf(
				"%w: LogLevel %d",
				ErrInvalidOpts,
				explicit,
			)
		}

		return explicit, nil
	}

	raw := strings.TrimSpace(get(EnvLogLevel))

	switch strings.ToLower(raw) {
	case "":
		return gormlogger.Warn, nil
	case "silent":
		return gormlogger.Silent, nil
	case "error":
		return gormlogger.Error, nil
	case "warn":
		return gormlogger.Warn, nil
	case "info":
		return gormlogger.Info, nil
	default:
		return 0, fmt.Errorf(
			"%w: %s=%q, want silent|error|warn|info",
			ErrInvalidEnv,
			EnvLogLevel,
			raw,
		)
	}
}
