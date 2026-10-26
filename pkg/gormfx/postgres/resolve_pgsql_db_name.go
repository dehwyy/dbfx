package postgres

import "strings"

func resolvePgsqlDBName(fallback string, get func(string) string) (string, error) {
	name := strings.TrimSpace(get(EnvPgsqlDBName))
	if name == "" {
		name = strings.TrimSpace(fallback)
	}

	if name == "" {
		return "", envError(
			EnvPgsqlDBName,
			"is empty and no database name was given",
		)
	}

	return name, nil
}
