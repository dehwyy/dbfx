package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dehwyy/dbfx/pkg/migrate"
)

func ParseBaseline(raw string) ([]migrate.BaselineStep, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parts := strings.Split(raw, ",")
	steps := make([]migrate.BaselineStep, 0, len(parts))

	for _, part := range parts {
		versionRaw, probe, found := strings.Cut(strings.TrimSpace(part), ":")
		if !found {
			return nil, fmt.Errorf("step %q: expected version:table[#column]", part)
		}

		version, err := strconv.ParseInt(strings.TrimSpace(versionRaw), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("step %q: %w", part, err)
		}

		step := migrate.BaselineStep{
			Version: version,
		}
		probe = strings.TrimSpace(probe)
		if probe != "" {
			step.Table, step.Column = splitProbe(probe)
		}

		steps = append(steps, step)
	}

	return steps, nil
}

func splitProbe(probe string) (string, string) {
	table, column, found := strings.Cut(probe, "#")
	if found {
		return table, column
	}

	return probe, ""
}
