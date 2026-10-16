package postgres

import "regexp"

var explainedPlaceholder = regexp.MustCompile(`\$(\d+)\$`)

func normalizePlaceholders(sql string) string {
	return explainedPlaceholder.ReplaceAllString(
		sql,
		"$$$1",
	)
}
