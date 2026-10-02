package migrate

import "fmt"

func planAdopt(
	steps []BaselineStep,
	known map[int64]struct{},
	present func(BaselineStep) (bool, error),
) ([]int64, error) {
	for _, step := range steps {
		if _, ok := known[step.Version]; !ok {
			return nil, fmt.Errorf(
				"%w: %d",
				ErrUnknownBaseline,
				step.Version,
			)
		}
	}

	versions := make(
		[]int64,
		0,
		len(steps),
	)
	for _, step := range steps {
		if step.Table != "" {
			ok, err := present(step)
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
		}
		versions = append(
			versions,
			step.Version,
		)
	}

	return versions, nil
}
