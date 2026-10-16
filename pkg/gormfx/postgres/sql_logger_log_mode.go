package postgres

import gormlogger "gorm.io/gorm/logger"

func (l *sqlLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	next := *l
	next.level = level

	return &next
}
