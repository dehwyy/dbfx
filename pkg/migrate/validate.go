package migrate

import (
	"errors"
	"fmt"
	"sort"
)

var (
	ErrNoFS            = errors.New("migrate: FS is required")
	ErrInvalidTable    = errors.New("migrate: invalid table name")
	ErrInvalidBaseline = errors.New("migrate: invalid baseline")
	ErrUnknownBaseline = errors.New("migrate: baseline version has no migration file")
)

func (c Config) validate() (Config, error) {
	if c.FS == nil {
		return c, ErrNoFS
	}

	if c.Table == "" {
		c.Table = DefaultTable
	}
	if !tableNamePattern.MatchString(c.Table) {
		return c, fmt.Errorf("%w: %q", ErrInvalidTable, c.Table)
	}

	baseline, err := normalizeBaseline(c.Baseline)
	if err != nil {
		return c, err
	}
	c.Baseline = baseline

	return c, nil
}

func normalizeBaseline(steps []BaselineStep) ([]BaselineStep, error) {
	if len(steps) == 0 {
		return nil, nil
	}

	sorted := make([]BaselineStep, len(steps))
	copy(sorted, steps)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Version < sorted[j].Version })

	for i, step := range sorted {
		if step.Version <= 0 {
			return nil, fmt.Errorf("%w: version must be positive, got %d", ErrInvalidBaseline, step.Version)
		}
		if i > 0 && sorted[i-1].Version == step.Version {
			return nil, fmt.Errorf("%w: duplicate version %d", ErrInvalidBaseline, step.Version)
		}
		if step.Column != "" && step.Table == "" {
			return nil, fmt.Errorf("%w: version %d has column without table", ErrInvalidBaseline, step.Version)
		}
		if step.Table != "" && !tableNamePattern.MatchString(step.Table) {
			return nil, fmt.Errorf("%w: version %d table %q", ErrInvalidBaseline, step.Version, step.Table)
		}
	}

	if sorted[0].Table == "" {
		return nil, fmt.Errorf("%w: first step %d needs a table probe", ErrInvalidBaseline, sorted[0].Version)
	}

	return sorted, nil
}
