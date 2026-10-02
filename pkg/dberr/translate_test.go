package dberr

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTranslate(t *testing.T) {
	plain := errors.New("boom")

	tests := []struct {
		name string
		in   error
		want error
	}{
		{
			name: "nil",
			in:   nil,
			want: nil,
		},
		{
			name: "record not found",
			in:   gorm.ErrRecordNotFound,
			want: ErrNotFound,
		},
		{
			name: "wrapped record not found",
			in: fmt.Errorf(
				"repo: %w",
				gorm.ErrRecordNotFound,
			),
			want: ErrNotFound,
		},
		{
			name: "gorm duplicated key",
			in:   gorm.ErrDuplicatedKey,
			want: ErrConflict,
		},
		{
			name: "gorm foreign key",
			in:   gorm.ErrForeignKeyViolated,
			want: ErrReference,
		},
		{
			name: "pg unique",
			in: &pgconn.PgError{
				Code: "23505",
			},
			want: ErrConflict,
		},
		{
			name: "pg foreign key",
			in: &pgconn.PgError{
				Code: "23503",
			},
			want: ErrReference,
		},
		{
			name: "pg not null",
			in: &pgconn.PgError{
				Code: "23502",
			},
			want: ErrNotNull,
		},
		{
			name: "pg check",
			in: &pgconn.PgError{
				Code: "23514",
			},
			want: ErrCheck,
		},
		{
			name: "pg serialization",
			in: &pgconn.PgError{
				Code: "40001",
			},
			want: ErrRetryable,
		},
		{
			name: "pg deadlock",
			in: &pgconn.PgError{
				Code: "40P01",
			},
			want: ErrRetryable,
		},
		{
			name: "wrapped pg unique",
			in: fmt.Errorf(
				"insert: %w",
				&pgconn.PgError{
					Code: "23505",
				},
			),
			want: ErrConflict,
		},
		{
			name: "unknown pg code passes through",
			in: &pgconn.PgError{
				Code: "42P01",
			},
			want: nil,
		},
		{
			name: "plain passes through",
			in:   plain,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				got := Translate(tt.in)

				if tt.in == nil {
					require.NoError(
						t,
						got,
					)
					return
				}

				require.ErrorIs(
					t,
					got,
					tt.in,
				)
				if tt.want == nil {
					require.Equal(
						t,
						tt.in,
						got,
					)
					return
				}
				require.ErrorIs(
					t,
					got,
					tt.want,
				)
			},
		)
	}
}

func TestTranslateKeepsSentinelsDistinct(t *testing.T) {
	got := Translate(gorm.ErrRecordNotFound)

	require.NotErrorIs(
		t,
		got,
		ErrConflict,
	)
}
