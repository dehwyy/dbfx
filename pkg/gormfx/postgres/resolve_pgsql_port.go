package postgres

import (
	"strconv"
	"strings"
)

func resolvePgsqlPort(get func(string) string) (int, error) {
	raw := strings.TrimSpace(get(EnvPgsqlPort))
	if raw == "" {
		return DefaultPgsqlPort, nil
	}

	port, err := strconv.Atoi(raw)
	if err != nil {
		return 0, envError(
			EnvPgsqlPort,
			"must be an integer",
		)
	}

	if port < 1 || port > 65535 {
		return 0, envError(
			EnvPgsqlPort,
			"must be in range 1..65535",
		)
	}

	return port, nil
}
