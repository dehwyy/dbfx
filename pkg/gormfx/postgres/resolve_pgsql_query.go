package postgres

import (
	"net/url"
	"strings"
)

func resolvePgsqlQuery(get func(string) string) (url.Values, error) {
	query := url.Values{}

	sslMode := strings.TrimSpace(get(EnvPgsqlSSLMode))
	if sslMode != "" {
		switch sslMode {
		case "disable", "allow", "prefer", "require", "verify-ca", "verify-full":
		default:
			return nil, envError(
				EnvPgsqlSSLMode,
				"must be one of disable, allow, prefer, require, verify-ca, verify-full",
			)
		}

		query.Set("sslmode", sslMode)
	}

	rootCert := strings.TrimSpace(get(EnvPgsqlSSLRootCert))
	if rootCert != "" {
		query.Set("sslrootcert", rootCert)
	}

	return query, nil
}
