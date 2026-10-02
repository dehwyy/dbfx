package fakesql

import "context"

func (c *fakeConn) Ping(ctx context.Context) error {
	switch c.mode {
	case "fail":
		return ErrPing
	case "hang":
		<-ctx.Done()
		return ctx.Err()
	}

	return nil
}
