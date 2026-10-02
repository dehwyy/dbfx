package gormfx

import (
	"context"
	"testing"

	"github.com/dehwyy/dbfx/internal/fakesql"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
	gormpg "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type testConfig struct {
	Opts Opts
}

func TestNewRequiresDatabase(t *testing.T) {
	_, err := New(Opts{})()

	require.Error(
		t,
		err,
	)
}

func TestProvideClosesOnStop(t *testing.T) {
	sqlDB := fakesql.Open(
		t,
		"ok",
	)
	lc := fxtest.NewLifecycle(t)

	db, err := provide(
		lc,
		func() (*gorm.DB, error) {
			return gorm.Open(
				gormpg.New(gormpg.Config{
					Conn: sqlDB,
				}),
				&gorm.Config{
					DisableAutomaticPing: true,
				},
			)
		},
	)
	require.NoError(
		t,
		err,
	)
	require.NotNil(
		t,
		db,
	)

	lc.RequireStart()
	require.NoError(
		t,
		sqlDB.PingContext(context.Background()),
	)

	lc.RequireStop()
	require.Error(
		t,
		sqlDB.PingContext(context.Background()),
	)
}

func TestProvideReturnsOpenError(t *testing.T) {
	lc := fxtest.NewLifecycle(t)

	_, err := provide(
		lc,
		New(Opts{}),
	)

	require.Error(
		t,
		err,
	)
}

func TestModuleWiresConfig(t *testing.T) {
	tests := []struct {
		name string
		opts Opts
	}{
		{
			name: "config without postgres fails startup graph",
			opts: Opts{},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				app := fx.New(
					fx.NopLogger,
					fx.Provide(func() *testConfig {
						return &testConfig{
							Opts: tt.opts,
						}
					}),
					Module(func(cfg *testConfig) Opts { return cfg.Opts }),
					fx.Invoke(func(*gorm.DB) {}),
				)

				require.ErrorContains(
					t,
					app.Err(),
					"no database provided",
				)
			},
		)
	}
}

func TestModuleOptsWires(t *testing.T) {
	app := fx.New(
		fx.NopLogger,
		ModuleOpts(Opts{}),
		fx.Invoke(func(*gorm.DB) {}),
	)

	require.ErrorContains(
		t,
		app.Err(),
		"no database provided",
	)
}
