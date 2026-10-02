package migrate

import (
	"context"
	"fmt"
)

func (m *Migrator) Up(ctx context.Context) ([]Result, error) {
	if len(m.cfg.Baseline) > 0 {
		if _, err := m.Adopt(ctx); err != nil {
			return nil, fmt.Errorf(
				"adopt baseline: %w",
				err,
			)
		}
	}

	applied, err := m.provider.Up(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"goose up: %w",
			err,
		)
	}

	results := make(
		[]Result,
		0,
		len(applied),
	)
	for _, item := range applied {
		results = append(
			results,
			Result{
				Version:  item.Source.Version,
				Path:     item.Source.Path,
				Duration: item.Duration,
			},
		)
	}

	return results, nil
}
