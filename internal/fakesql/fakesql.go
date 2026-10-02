package fakesql

import (
	"errors"
	"sync"
)

var (
	ErrPing = errors.New("fake ping failed")

	registerOnce sync.Once
)

const driverName = "dbfx-fake"
