package postgres

import "unicode/utf8"

func truncateSQL(sql string, limit int) string {
	if len(sql) <= limit {
		return sql
	}

	cut := limit
	for cut > 0 && !utf8.RuneStart(sql[cut]) {
		cut--
	}

	return sql[:cut]
}
