package migrate

func (m *Migrator) Versions() []int64 {
	sources := m.provider.ListSources()

	versions := make([]int64, 0, len(sources))
	for _, source := range sources {
		versions = append(versions, source.Version)
	}

	return versions
}
