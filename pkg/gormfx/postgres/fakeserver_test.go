package postgres

import (
	"bytes"
	"fmt"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"
	gormlogger "gorm.io/gorm/logger"
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

func TestNewInstallsSQLLogger(t *testing.T) {
	dsn := startFakePG(t)
	out := &bytes.Buffer{}

	t.Setenv(
		EnvLogLevel,
		"info",
	)
	t.Setenv(
		EnvLogSlowThreshold,
		"1s",
	)

	conn, err := New(Opts{
		ConnectionStrings: []string{dsn},
		PingTimeout:       5 * time.Second,
		LogOutput:         out,
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

	installed, ok := conn.Logger.(*sqlLogger)
	require.True(
		t,
		ok,
	)
	require.Equal(
		t,
		gormlogger.Info,
		installed.level,
	)
	require.Equal(
		t,
		time.Second,
		installed.slow,
	)
}

func TestNewRejectsBrokenLogEnv(t *testing.T) {
	tests := []struct {
		name string
		env  string
		val  string
	}{
		{
			name: "level",
			env:  EnvLogLevel,
			val:  "loud",
		},
		{
			name: "threshold",
			env:  EnvLogSlowThreshold,
			val:  "soon",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				dsn := startFakePG(t)
				t.Setenv(
					tt.env,
					tt.val,
				)

				conn, err := New(Opts{
					ConnectionStrings: []string{dsn},
					PingTimeout:       5 * time.Second,
				})

				require.ErrorIs(
					t,
					err,
					ErrInvalidEnv,
				)
				require.Nil(
					t,
					conn,
				)
			},
		)
	}
}
