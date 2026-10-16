package postgres

import (
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

const gormFramePrefix = "gorm.io/"

var loggerFramePrefix = reflect.TypeFor[sqlLogger]().PkgPath() + ".(*sqlLogger)."

func callerFile() string {
	pcs := [24]uintptr{}
	count := runtime.Callers(
		2,
		pcs[:],
	)
	frames := runtime.CallersFrames(pcs[:count])

	for {
		frame, more := frames.Next()
		if frame.File != "" && !isInternalFrame(frame.Function) {
			return frame.File + ":" + strconv.Itoa(frame.Line)
		}
		if !more {
			return ""
		}
	}
}

func isInternalFrame(function string) bool {
	return strings.HasPrefix(
		function,
		gormFramePrefix,
	) || strings.HasPrefix(
		function,
		loggerFramePrefix,
	)
}
