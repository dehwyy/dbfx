package fakesql

func (c *fakeConn) Close() error {
	return nil
}
