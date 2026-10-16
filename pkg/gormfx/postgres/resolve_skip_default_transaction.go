package postgres

import (
	"fmt"
	"strconv"
	"strings"
)

func resolveSkipDefaultTransaction(explicit bool, get func(string) string) (bool, error) {
	raw := strings.TrimSpace(get(EnvSkipDefaultTransaction))
	if raw == "" {
		return explicit, nil
	}

	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		return false, fmt.Errorf(
			"%w: %s=%q: %w",
			ErrInvalidEnv,
			EnvSkipDefaultTransaction,
			raw,
			err,
		)
	}

	return explicit || enabled, nil
}
