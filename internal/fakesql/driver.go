package fakesql

import "database/sql/driver"

type fakeDriver struct{}

func (fakeDriver) Open(name string) (driver.Conn, error) {
	return &fakeConn{
		mode: name,
	}, nil
}
