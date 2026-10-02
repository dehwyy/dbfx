package postgres

import (
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"
)

func startFakePG(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen(
		"tcp",
		"127.0.0.1:0",
	)
	require.NoError(
		t,
		err,
	)
	t.Cleanup(func() {
		require.NoError(
			t,
			listener.Close(),
		)
	})

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go serveFakePG(conn)
		}
	}()

	addr := listener.Addr().(*net.TCPAddr)

	return fmt.Sprintf(
		"host=127.0.0.1 port=%d user=u password=p dbname=d sslmode=disable",
		addr.Port,
	)
}

func serveFakePG(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Debug(
				"fake pg close",
				"err",
				err,
			)
		}
	}()

	backend := pgproto3.NewBackend(
		conn,
		conn,
	)

	if _, err := backend.ReceiveStartupMessage(); err != nil {
		return
	}

	backend.Send(&pgproto3.AuthenticationOk{})
	backend.Send(&pgproto3.ReadyForQuery{
		TxStatus: 'I',
	})
	if err := backend.Flush(); err != nil {
		return
	}

	for {
		msg, err := backend.Receive()
		if err != nil {
			return
		}

		switch msg.(type) {
		case *pgproto3.Query:
			backend.Send(&pgproto3.EmptyQueryResponse{})
			backend.Send(&pgproto3.ReadyForQuery{
				TxStatus: 'I',
			})
			if err := backend.Flush(); err != nil {
				return
			}
		case *pgproto3.Terminate:
			return
		}
	}
}

func TestNewAppliesPoolLimitsWithSingleDSN(t *testing.T) {
	dsn := startFakePG(t)

	conn, err := New(Opts{
		ConnectionStrings: []string{dsn},
		ConnectionMaxOpen: 5,
		ConnectionMaxIdle: 2,
		PingTimeout:       5 * time.Second,
	})
	require.NoError(
		t,
		err,
	)

	sqlDB, err := conn.DB()
	require.NoError(
		t,
		err,
	)
	t.Cleanup(func() {
		require.NoError(
			t,
			sqlDB.Close(),
		)
	})

	require.Equal(
		t,
		5,
		sqlDB.Stats().MaxOpenConnections,
	)
}
