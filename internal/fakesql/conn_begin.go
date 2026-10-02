package fakesql

import "database/sql/driver"

func (c *fakeConn) Begin() (driver.Tx, error) {
	return nil, driver.ErrSkip
}
