package postgres

import "os"

func ConnectionStringFromEnv(dbName string) (string, bool, error) {
	return connectionStringFromEnv(
		dbName,
		os.Getenv,
	)
}
