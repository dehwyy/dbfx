package postgres

import "fmt"

func envError(name string, reason string) error {
	return fmt.Errorf(
		"%w: %s %s",
		ErrInvalidEnv,
		name,
		reason,
	)
}
