package postgres

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/dehwyy/dbfx/internal/fakesql"
	"github.com/stretchr/testify/require"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const secretValue = "super-secret-value"

type wrappedLogger struct {
	gormlogger.Interface
}

func (l wrappedLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return wrappedLogger{
		Interface: l.Interface.LogMode(level),
	}
}

func (l wrappedLogger) ParamsFilter(_ context.Context, sql string, _ ...any) (string, []any) {
	return sql, nil
}

func newTestLogger(level gormlogger.LogLevel, slow time.Duration) (*sqlLogger, *bytes.Buffer) {
	out := &bytes.Buffer{}

	return newSQLLogger(logSettings{
		level:  level,
		slow:   slow,
		output: out,
	}), out
}

func decodeLines(t *testing.T, out *bytes.Buffer) []map[string]any {
	t.Helper()

	raw := strings.TrimSuffix(
		out.String(),
		"\n",
	)
	if raw == "" {
		return nil
	}

	lines := strings.Split(
		raw,
		"\n",
	)
	records := make(
		[]map[string]any,
		0,
		len(lines),
	)
	for _, line := range lines {
		record := map[string]any{}
		require.NoError(
			t,
			json.Unmarshal(
				[]byte(line),
				&record,
			),
			line,
		)
		records = append(
			records,
			record,
		)
	}

	return records
}

type traceProbe struct {
	calls int
	sql   string
	rows  int64
}

func (p *traceProbe) fc() (string, int64) {
	p.calls++

	return p.sql, p.rows
}

func TestSQLLoggerTrace(t *testing.T) {
	tests := []struct {
		name      string
		level     gormlogger.LogLevel
		slow      time.Duration
		elapsed   time.Duration
		err       error
		sql       string
		rows      int64
		wantLines int
		wantLevel string
		wantMsg   string
		wantError string
	}{
		{
			name:      "not found is not logged at warn",
			level:     gormlogger.Warn,
			slow:      200 * time.Millisecond,
			err:       gormlogger.ErrRecordNotFound,
			sql:       "SELECT 1",
			wantLines: 0,
		},
		{
			name:  "wrapped not found is not logged at warn",
			level: gormlogger.Warn,
			slow:  200 * time.Millisecond,
			err: errors.Join(
				errors.New("ctx"),
				gormlogger.ErrRecordNotFound,
			),
			sql:       "SELECT 1",
			wantLines: 0,
		},
		{
			name:      "fast query without error is not logged",
			level:     gormlogger.Warn,
			slow:      200 * time.Millisecond,
			sql:       "SELECT 1",
			wantLines: 0,
		},
		{
			name:      "slow query is one warn line",
			level:     gormlogger.Warn,
			slow:      200 * time.Millisecond,
			elapsed:   time.Second,
			sql:       "SELECT *\nFROM orders\nWHERE idempotency_key = $1$",
			rows:      3,
			wantLines: 1,
			wantLevel: "warn",
			wantMsg:   "slow sql",
		},
		{
			name:      "slow query is not logged at error level",
			level:     gormlogger.Error,
			slow:      200 * time.Millisecond,
			elapsed:   time.Second,
			sql:       "SELECT 1",
			wantLines: 0,
		},
		{
			name:      "slow detection disabled by zero threshold",
			level:     gormlogger.Warn,
			slow:      0,
			elapsed:   time.Hour,
			sql:       "SELECT 1",
			wantLines: 0,
		},
		{
			name:      "unique violation is logged",
			level:     gormlogger.Warn,
			slow:      200 * time.Millisecond,
			err:       gorm.ErrDuplicatedKey,
			sql:       "INSERT INTO orders (key) VALUES ($1$)",
			rows:      0,
			wantLines: 1,
			wantLevel: "error",
			wantMsg:   "sql error",
			wantError: gorm.ErrDuplicatedKey.Error(),
		},
		{
			name:      "error level logs errors only",
			level:     gormlogger.Error,
			slow:      200 * time.Millisecond,
			err:       errors.New("boom"),
			sql:       "UPDATE t SET a = $1$",
			rows:      -1,
			wantLines: 1,
			wantLevel: "error",
			wantMsg:   "sql error",
			wantError: "boom",
		},
		{
			name:      "silent logs nothing",
			level:     gormlogger.Silent,
			slow:      time.Nanosecond,
			elapsed:   time.Hour,
			err:       errors.New("boom"),
			sql:       "SELECT 1",
			wantLines: 0,
		},
		{
			name:      "info logs fast query",
			level:     gormlogger.Info,
			slow:      200 * time.Millisecond,
			sql:       "SELECT 1",
			rows:      1,
			wantLines: 1,
			wantLevel: "info",
			wantMsg:   "sql",
		},
		{
			name:      "info logs not found without error field",
			level:     gormlogger.Info,
			slow:      200 * time.Millisecond,
			err:       gormlogger.ErrRecordNotFound,
			sql:       "SELECT 1",
			rows:      0,
			wantLines: 1,
			wantLevel: "info",
			wantMsg:   "sql",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				l, out := newTestLogger(
					tt.level,
					tt.slow,
				)
				probe := &traceProbe{
					sql:  tt.sql,
					rows: tt.rows,
				}

				l.Trace(
					context.Background(),
					time.Now().Add(-tt.elapsed),
					probe.fc,
					tt.err,
				)

				records := decodeLines(
					t,
					out,
				)
				require.Len(
					t,
					records,
					tt.wantLines,
				)
				if tt.wantLines == 0 {
					require.Zero(
						t,
						probe.calls,
					)
					return
				}

				require.Equal(
					t,
					1,
					probe.calls,
				)
				record := records[0]
				require.Equal(
					t,
					tt.wantLevel,
					record["level"],
				)
				require.Equal(
					t,
					tt.wantMsg,
					record["message"],
				)
				require.NotEmpty(
					t,
					record["time"],
				)
				require.NotContains(
					t,
					record,
					"msg",
				)
				require.Contains(
					t,
					record,
					"duration_ms",
				)
				require.Contains(
					t,
					record["caller"],
					"sql_logger_test.go:",
				)
				require.Equal(
					t,
					1,
					strings.Count(
						out.String(),
						"\n",
					),
				)
				require.NotContains(
					t,
					out.String(),
					"$1$",
				)

				if tt.rows >= 0 {
					require.EqualValues(
						t,
						tt.rows,
						record["rows"],
					)
				} else {
					require.NotContains(
						t,
						record,
						"rows",
					)
				}

				if tt.wantError == "" {
					require.NotContains(
						t,
						record,
						"error",
					)
					return
				}
				require.Equal(
					t,
					tt.wantError,
					record["error"],
				)
			},
		)
	}
}

func TestSQLLoggerTraceKeepsPlaceholders(t *testing.T) {
	l, out := newTestLogger(
		gormlogger.Warn,
		200*time.Millisecond,
	)
	probe := &traceProbe{
		sql: "SELECT * FROM orders WHERE a = $1$ AND b = $2$ AND c = $10$",
	}

	l.Trace(
		context.Background(),
		time.Now().Add(-time.Second),
		probe.fc,
		nil,
	)

	records := decodeLines(
		t,
		out,
	)
	require.Len(
		t,
		records,
		1,
	)
	require.Equal(
		t,
		"SELECT * FROM orders WHERE a = $1 AND b = $2 AND c = $10",
		records[0]["sql"],
	)
}

func TestSQLLoggerTraceTruncatesSQL(t *testing.T) {
	long := "SELECT '" + strings.Repeat(
		"я",
		MaxLoggedSQL,
	) + "'"

	tests := []struct {
		name     string
		sql      string
		wantCut  bool
		wantSize int
	}{
		{
			name: "short sql is intact",
			sql:  "SELECT 1",
		},
		{
			name: "exact limit is intact",
			sql: strings.Repeat(
				"a",
				MaxLoggedSQL,
			),
		},
		{
			name: "ascii over limit is cut",
			sql: strings.Repeat(
				"a",
				MaxLoggedSQL+500,
			),
			wantCut:  true,
			wantSize: MaxLoggedSQL + 500,
		},
		{
			name:     "multibyte is cut on a rune boundary",
			sql:      long,
			wantCut:  true,
			wantSize: len(long),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				l, out := newTestLogger(
					gormlogger.Warn,
					200*time.Millisecond,
				)
				probe := &traceProbe{
					sql: tt.sql,
				}

				l.Trace(
					context.Background(),
					time.Now().Add(-time.Second),
					probe.fc,
					nil,
				)

				records := decodeLines(
					t,
					out,
				)
				require.Len(
					t,
					records,
					1,
				)
				logged, ok := records[0]["sql"].(string)
				require.True(
					t,
					ok,
				)
				require.LessOrEqual(
					t,
					len(logged),
					MaxLoggedSQL,
				)
				require.NotContains(
					t,
					logged,
					"�",
				)

				if !tt.wantCut {
					require.Equal(
						t,
						tt.sql,
						logged,
					)
					require.NotContains(
						t,
						records[0],
						"sql_len",
					)
					return
				}
				require.EqualValues(
					t,
					tt.wantSize,
					records[0]["sql_len"],
				)
			},
		)
	}
}

func TestSQLLoggerLogMode(t *testing.T) {
	l, out := newTestLogger(
		gormlogger.Warn,
		200*time.Millisecond,
	)

	silent := l.LogMode(gormlogger.Silent)
	info := l.LogMode(gormlogger.Info)

	require.Equal(
		t,
		gormlogger.Warn,
		l.level,
	)
	require.IsType(
		t,
		&sqlLogger{},
		silent,
	)
	_, ok := silent.(gorm.ParamsFilter)
	require.True(
		t,
		ok,
	)

	probe := &traceProbe{
		sql: "SELECT 1",
	}
	silent.Trace(
		context.Background(),
		time.Now().Add(-time.Second),
		probe.fc,
		errors.New("boom"),
	)
	require.Empty(
		t,
		out.String(),
	)

	info.Trace(
		context.Background(),
		time.Now(),
		probe.fc,
		nil,
	)
	require.Len(
		t,
		decodeLines(
			t,
			out,
		),
		1,
	)

	l.Trace(
		context.Background(),
		time.Now(),
		probe.fc,
		nil,
	)
	require.Len(
		t,
		decodeLines(
			t,
			out,
		),
		1,
	)
}

func TestSQLLoggerInfoWarnError(t *testing.T) {
	tests := []struct {
		name      string
		level     gormlogger.LogLevel
		call      func(l gormlogger.Interface)
		wantLevel string
		wantMsg   string
	}{
		{
			name:  "info at info level",
			level: gormlogger.Info,
			call: func(l gormlogger.Interface) {
				l.Info(
					context.Background(),
					"hello %s",
					"world",
				)
			},
			wantLevel: "info",
			wantMsg:   "hello world",
		},
		{
			name:  "info is dropped at warn level",
			level: gormlogger.Warn,
			call: func(l gormlogger.Interface) {
				l.Info(
					context.Background(),
					"hello",
				)
			},
		},
		{
			name:  "warn at warn level",
			level: gormlogger.Warn,
			call: func(l gormlogger.Interface) {
				l.Warn(
					context.Background(),
					"100% careful",
				)
			},
			wantLevel: "warn",
			wantMsg:   "100% careful",
		},
		{
			name:  "warn is dropped at error level",
			level: gormlogger.Error,
			call: func(l gormlogger.Interface) {
				l.Warn(
					context.Background(),
					"hello",
				)
			},
		},
		{
			name:  "error at error level",
			level: gormlogger.Error,
			call: func(l gormlogger.Interface) {
				l.Error(
					context.Background(),
					"failed: %v",
					errors.New("boom"),
				)
			},
			wantLevel: "error",
			wantMsg:   "failed: boom",
		},
		{
			name:  "error is dropped at silent level",
			level: gormlogger.Silent,
			call: func(l gormlogger.Interface) {
				l.Error(
					context.Background(),
					"hello",
				)
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				l, out := newTestLogger(
					tt.level,
					0,
				)

				tt.call(l)

				records := decodeLines(
					t,
					out,
				)
				if tt.wantMsg == "" {
					require.Empty(
						t,
						records,
					)
					return
				}
				require.Len(
					t,
					records,
					1,
				)
				require.Equal(
					t,
					tt.wantLevel,
					records[0]["level"],
				)
				require.Equal(
					t,
					tt.wantMsg,
					records[0]["message"],
				)
				require.Contains(
					t,
					records[0]["caller"],
					"sql_logger_test.go:",
				)
			},
		)
	}
}

func openDryRun(t *testing.T, l gormlogger.Interface) *gorm.DB {
	t.Helper()

	sqlDB := fakesql.Open(
		t,
		"ok",
	)
	t.Cleanup(func() {
		require.NoError(
			t,
			sqlDB.Close(),
		)
	})

	db, err := gorm.Open(
		gormpg.New(gormpg.Config{
			Conn: sqlDB,
		}),
		&gorm.Config{
			Logger:               l,
			DryRun:               true,
			DisableAutomaticPing: true,
		},
	)
	require.NoError(
		t,
		err,
	)

	return db
}

func TestSQLLoggerHidesValuesThroughGorm(t *testing.T) {
	tests := []struct {
		name string
		wrap func(l gormlogger.Interface) gormlogger.Interface
	}{
		{
			name: "plain logger",
			wrap: func(l gormlogger.Interface) gormlogger.Interface { return l },
		},
		{
			name: "wrapped logger with params filter",
			wrap: func(l gormlogger.Interface) gormlogger.Interface {
				return wrappedLogger{
					Interface: l,
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				l, out := newTestLogger(
					gormlogger.Info,
					200*time.Millisecond,
				)
				db := openDryRun(
					t,
					tt.wrap(l),
				)

				rows := []map[string]any{}
				require.NoError(
					t,
					db.Table("orders").
						Where(
							"api_key_id = ? AND amount > ?",
							secretValue,
							42,
						).
						Find(&rows).
						Error,
				)

				require.NotContains(
					t,
					out.String(),
					secretValue,
				)
				records := decodeLines(
					t,
					out,
				)
				require.Len(
					t,
					records,
					1,
				)
				require.Contains(
					t,
					records[0]["sql"],
					"api_key_id = $1 AND amount > $2",
				)
				require.Contains(
					t,
					records[0]["caller"],
					"sql_logger_test.go:",
				)
				require.NotContains(
					t,
					records[0]["caller"],
					"gorm.io",
				)
			},
		)
	}
}

func TestSQLLoggerDebugAndSessionKeepWorking(t *testing.T) {
	l, out := newTestLogger(
		gormlogger.Warn,
		200*time.Millisecond,
	)
	db := openDryRun(
		t,
		l,
	)

	rows := []map[string]any{}
	require.NoError(
		t,
		db.Table("orders").Where(
			"k = ?",
			secretValue,
		).Find(&rows).Error,
	)
	require.Empty(
		t,
		out.String(),
	)

	require.NoError(
		t,
		db.Debug().Table("orders").Where(
			"k = ?",
			secretValue,
		).Find(&rows).Error,
	)
	require.Len(
		t,
		decodeLines(
			t,
			out,
		),
		1,
	)

	silent, silentOut := newTestLogger(
		gormlogger.Info,
		0,
	)
	require.NoError(
		t,
		db.Session(&gorm.Session{
			Logger: silent.LogMode(gormlogger.Silent),
		}).Table("orders").Where(
			"k = ?",
			secretValue,
		).Find(&rows).Error,
	)
	require.Empty(
		t,
		silentOut.String(),
	)
	require.NotContains(
		t,
		out.String(),
		secretValue,
	)
}
