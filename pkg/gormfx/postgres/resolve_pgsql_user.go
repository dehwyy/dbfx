package postgres

import "strings"

func resolvePgsqlUser(get func(string) string) (string, error) {
	user := strings.TrimSpace(get(EnvPgsqlUser))
	username := strings.TrimSpace(get(EnvPgsqlUsername))

	if user != "" && username != "" && user != username {
		return "", envError(
			EnvPgsqlUsername,
			"differs from "+EnvPgsqlUser,
		)
	}

	if user == "" {
		user = username
	}

	if user == "" {
		return "", envError(
			EnvPgsqlUser,
			"and "+EnvPgsqlUsername+" are both empty",
		)
	}

	return user, nil
}
