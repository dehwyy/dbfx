package migrate

import (
	"io/fs"
	"log/slog"
	"regexp"

	"github.com/pressly/goose/v3"
)

const DefaultTable = "goose_db_version"

var tableNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)?$`)

type BaselineStep struct {
	Version int64
	Table   string
	Column  string
}

type Config struct {
	FS           fs.FS
	Dir          string
	Table        string
	LockID       int64
	Baseline     []BaselineStep
	GoMigrations []*goose.Migration
	Logger       *slog.Logger
}
