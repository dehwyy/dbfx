package fakesql

import "database/sql/driver"

func (c *fakeConn) Prepare(string) (driver.Stmt, error) {
	return nil, driver.ErrSkip
}
