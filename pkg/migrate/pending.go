package migrate

import "context"

func (m *Migrator) Pending(ctx context.Context) (bool, error) {
	return m.provider.HasPending(ctx)
}
