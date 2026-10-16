package postgres

import (
	"fmt"
	"strings"
	"time"
)

func resolveLogSlowThreshold(explicit time.Duration, get func(string) string) (time.Duration, error) {
	switch {
	case explicit > 0:
		return explicit, nil
	case explicit < 0:
		return 0, nil
	}

	raw := strings.TrimSpace(get(EnvLogSlowThreshold))
	if raw == "" {
		return DefaultLogSlowThreshold, nil
	}

	threshold, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf(
			"%w: %s=%q: %w",
			ErrInvalidEnv,
			EnvLogSlowThreshold,
			raw,
			err,
		)
	}
	if threshold < 0 {
		return 0, fmt.Errorf(
			"%w: %s=%q, want >= 0",
			ErrInvalidEnv,
			EnvLogSlowThreshold,
			raw,
		)
	}

	return threshold, nil
}
