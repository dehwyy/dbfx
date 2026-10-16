package postgres

import (
	"io"
	"time"

	gormlogger "gorm.io/gorm/logger"
)

type logSettings struct {
	level  gormlogger.LogLevel
	slow   time.Duration
	output io.Writer
}
