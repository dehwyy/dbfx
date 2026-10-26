package postgres

import (
	"net"
	"net/url"
	"strconv"
	"strings"
)

func connectionStringFromEnv(dbName string, get func(string) string) (string, bool, error) {
	host := strings.TrimSpace(get(EnvPgsqlHost))
	if host == "" {
		return "", false, nil
	}

	port, err := resolvePgsqlPort(get)
	if err != nil {
		return "", false, err
	}

	database, err := resolvePgsqlDBName(
		dbName,
		get,
	)
	if err != nil {
		return "", false, err
	}

	user, err := resolvePgsqlUser(get)
	if err != nil {
		return "", false, err
	}

	password := get(EnvPgsqlPassword)
	if password == "" {
		return "", false, envError(
			EnvPgsqlPassword,
			"must not be empty",
		)
	}

	query, err := resolvePgsqlQuery(get)
	if err != nil {
		return "", false, err
	}

	dsn := url.URL{
		Scheme: "postgres",
		User: url.UserPassword(
			user,
			password,
		),
		Host: net.JoinHostPort(
			host,
			strconv.Itoa(port),
		),
		Path:     "/" + database,
		RawQuery: query.Encode(),
	}

	return dsn.String(), true, nil
}
