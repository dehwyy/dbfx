package migrate

import (
	"context"
	"errors"
	"os"
	"testing"
	"testing/fstest"
	"time"

	"github.com/dehwyy/dbfx/internal/fakesql"
	"github.com/stretchr/testify/require"
	"go.uber.org/fx"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		cfg       Config
		wantErr   error
		wantTable string
	}{
		{
			name:    "no fs",
			cfg:     Config{},
			wantErr: ErrNoFS,
		},
		{
			name: "default table",
			cfg: Config{
				FS: fstest.MapFS{},
			},
			wantTable: DefaultTable,
		},
		{
			name: "custom table",
			cfg: Config{
				FS:    fstest.MapFS{},
				Table: "app.schema_versions",
			},
			wantTable: "app.schema_versions",
		},
		{
			name: "injection in table",
			cfg: Config{
				FS:    fstest.MapFS{},
				Table: "t; DROP TABLE x",
			},
			wantErr: ErrInvalidTable,
		},
		{
			name: "baseline first without probe",
			cfg: Config{
				FS: fstest.MapFS{},
				Baseline: []BaselineStep{
					{
						Version: 1,
					},
				},
			},
			wantErr: ErrInvalidBaseline,
		},
		{
			name: "baseline zero version",
			cfg: Config{
				FS: fstest.MapFS{},
				Baseline: []BaselineStep{
					{
						Version: 0,
						Table:   "a",
					},
				},
			},
			wantErr: ErrInvalidBaseline,
		},
		{
			name: "baseline duplicate",
			cfg: Config{
				FS: fstest.MapFS{},
				Baseline: []BaselineStep{
					{
						Version: 1,
						Table:   "a",
					},
					{
						Version: 1,
						Table:   "b",
					},
				},
			},
			wantErr: ErrInvalidBaseline,
		},
		{
			name: "baseline column without table",
			cfg: Config{
				FS: fstest.MapFS{},
				Baseline: []BaselineStep{
					{
						Version: 1,
						Table:   "a",
					},
					{
						Version: 2,
						Column:  "c",
					},
				},
			},
			wantErr: ErrInvalidBaseline,
		},
		{
			name: "baseline bad table",
			cfg: Config{
				FS: fstest.MapFS{},
				Baseline: []BaselineStep{
					{
						Version: 1,
						Table:   "a b",
					},
				},
			},
			wantErr: ErrInvalidBaseline,
		},
		{
			name: "baseline ok unsorted",
			cfg: Config{
				FS: fstest.MapFS{},
				Baseline: []BaselineStep{
					{
						Version: 2,
						Table:   "a",
						Column:  "c",
					},
					{
						Version: 1,
						Table:   "a",
					},
				},
			},
			wantTable: DefaultTable,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				got, err := tt.cfg.validate()

				if tt.wantErr != nil {
					require.ErrorIs(
						t,
						err,
						tt.wantErr,
					)
					return
				}
				require.NoError(
					t,
					err,
				)
				require.Equal(
					t,
					tt.wantTable,
					got.Table,
				)
				for i := 1; i < len(got.Baseline); i++ {
					require.Less(
						t,
						got.Baseline[i-1].Version,
						got.Baseline[i].Version,
					)
				}
			},
		)
	}
}

func TestPlanAdopt(t *testing.T) {
	known := map[int64]struct{}{1: {}, 2: {}, 3: {}, 4: {}}
	steps := []BaselineStep{
		{
			Version: 1,
			Table:   "orders",
		},
		{
			Version: 2,
			Table:   "providers",
		},
		{
			Version: 3,
		},
		{
			Version: 4,
			Table:   "orders",
			Column:  "balance_phase",
		},
	}

	tests := []struct {
		name    string
		steps   []BaselineStep
		have    map[string]bool
		want    []int64
		wantErr error
	}{
		{
			name:  "fresh database",
			steps: steps,
			have:  map[string]bool{},
			want:  []int64{},
		},
		{
			name:  "fully legacy",
			steps: steps,
			have:  map[string]bool{"orders": true, "providers": true, "orders#balance_phase": true},
			want:  []int64{1, 2, 3, 4},
		},
		{
			name:  "stops at first missing probe",
			steps: steps,
			have:  map[string]bool{"orders": true},
			want:  []int64{1},
		},
		{
			name:  "gap in the middle stops at the gap",
			steps: steps,
			have:  map[string]bool{"orders": true, "orders#balance_phase": true},
			want:  []int64{1},
		},
		{
			name:  "partial legacy missing column",
			steps: steps,
			have:  map[string]bool{"orders": true, "providers": true},
			want:  []int64{1, 2, 3},
		},
		{
			name: "unknown version",
			steps: []BaselineStep{
				{
					Version: 9,
					Table:   "x",
				},
			},
			have:    map[string]bool{"x": true},
			wantErr: ErrUnknownBaseline,
		},
		{
			name:    "probe error",
			steps:   steps,
			have:    nil,
			wantErr: errProbe,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				got, err := planAdopt(
					tt.steps,
					known,
					func(step BaselineStep) (bool, error) {
						if tt.have == nil {
							return false, errProbe
						}
						key := step.Table
						if step.Column != "" {
							key += "#" + step.Column
						}
						return tt.have[key], nil
					},
				)

				if tt.wantErr != nil {
					require.ErrorIs(
						t,
						err,
						tt.wantErr,
					)
					return
				}
				require.NoError(
					t,
					err,
				)
				require.Equal(
					t,
					tt.want,
					got,
				)
			},
		)
	}
}

var errProbe = errors.New("probe failed")

func TestNewListsVersions(t *testing.T) {
	tests := []struct {
		name         string
		cfg          Config
		wantVersions []int64
		wantErr      bool
	}{
		{
			name: "dir fs",
			cfg: Config{
				FS: os.DirFS("testdata/ok"),
			},
			wantVersions: []int64{1, 2},
		},
		{
			name: "sub dir",
			cfg: Config{
				FS:  os.DirFS("testdata/nested"),
				Dir: "sql",
			},
			wantVersions: []int64{1},
		},
		{
			name: "missing sub dir",
			cfg: Config{
				FS:  os.DirFS("testdata/nested"),
				Dir: "nope",
			},
			wantErr: true,
		},
		{
			name: "invalid table",
			cfg: Config{
				FS:    os.DirFS("testdata/ok"),
				Table: "bad name",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				db := fakesql.Open(
					t,
					"ok",
				)
				t.Cleanup(func() {
					require.NoError(
						t,
						db.Close(),
					)
				})

				migrator, err := New(
					db,
					tt.cfg,
				)

				if tt.wantErr {
					require.Error(
						t,
						err,
					)
					return
				}
				require.NoError(
					t,
					err,
				)
				require.ElementsMatch(
					t,
					tt.wantVersions,
					migrator.Versions(),
				)
			},
		)
	}
}

func TestAdoptWithoutBaselineIsNoop(t *testing.T) {
	db := fakesql.Open(
		t,
		"ok",
	)
	t.Cleanup(func() {
		require.NoError(
			t,
			db.Close(),
		)
	})

	migrator, err := New(
		db,
		Config{
			FS: os.DirFS("testdata/ok"),
		},
	)
	require.NoError(
		t,
		err,
	)

	adopted, err := migrator.Adopt(context.Background())

	require.NoError(
		t,
		err,
	)
	require.Empty(
		t,
		adopted,
	)
}

type fakeRunner struct {
	results []Result
	err     error
	calls   int
}

func (f *fakeRunner) Up(context.Context) ([]Result, error) {
	f.calls++
	return f.results, f.err
}

func TestRunOnce(t *testing.T) {
	tests := []struct {
		name     string
		runner   *fakeRunner
		wantCode int
	}{
		{
			name: "success shuts down cleanly",
			runner: &fakeRunner{
				results: []Result{
					{
						Version: 1,
						Path:    "00001_init.sql",
					},
				},
			},
			wantCode: 0,
		},
		{
			name: "failure shuts down with exit code",
			runner: &fakeRunner{
				err: errProbe,
			},
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name,
			func(t *testing.T) {
				app := fx.New(
					fx.NopLogger,
					fx.Invoke(func(lc fx.Lifecycle, shutdowner fx.Shutdowner) {
						runOnce(
							lc,
							tt.runner,
							shutdowner,
						)
					}),
				)

				require.NoError(
					t,
					app.Start(context.Background()),
				)

				select {
				case signal := <-app.Wait():
					require.Equal(
						t,
						tt.wantCode,
						signal.ExitCode,
					)
				case <-time.After(2 * time.Second):
					t.Fatal("shutdown not requested")
				}

				require.NoError(
					t,
					app.Stop(context.Background()),
				)
				require.Equal(
					t,
					1,
					tt.runner.calls,
				)
			},
		)
	}
}
