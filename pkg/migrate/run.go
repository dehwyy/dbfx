package migrate

import (
	"context"
	"log/slog"

	"go.uber.org/fx"
)

type upRunner interface {
	Up(ctx context.Context) ([]Result, error)
}

func RunOnce() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, migrator *Migrator, shutdowner fx.Shutdowner) {
			runOnce(lc, migrator, shutdowner)
		},
	)
}

func runOnce(lc fx.Lifecycle, runner upRunner, shutdowner fx.Shutdowner) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	lc.Append(
		fx.Hook{
			OnStart: func(context.Context) error {
				go func() {
					defer close(done)
					execute(ctx, runner, shutdowner)
				}()

				return nil
			},
			OnStop: func(stopCtx context.Context) error {
				cancel()

				select {
				case <-done:
				case <-stopCtx.Done():
					return stopCtx.Err()
				}

				return nil
			},
		},
	)
}

func execute(ctx context.Context, runner upRunner, shutdowner fx.Shutdowner) {
	results, err := runner.Up(ctx)
	if err != nil {
		slog.Error("migrate failed", slog.Any("error", err))
		if shutdownErr := shutdowner.Shutdown(fx.ExitCode(1)); shutdownErr != nil {
			slog.Error("shutdown failed", slog.Any("error", shutdownErr))
		}
		return
	}

	for _, result := range results {
		slog.Info(
			"migration applied",
			slog.Int64("version", result.Version),
			slog.String("path", result.Path),
			slog.Duration("duration", result.Duration),
		)
	}
	slog.Info("migrate done", slog.Int("applied", len(results)))

	if shutdownErr := shutdowner.Shutdown(); shutdownErr != nil {
		slog.Error("shutdown failed", slog.Any("error", shutdownErr))
	}
}
