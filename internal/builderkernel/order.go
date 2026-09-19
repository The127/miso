package builderkernel

// order is the order to load the wanted modules in, each of them after the
// modules it depends on.
func order(have []info, want ...string) []string {
	loaded := make([]string, 0, len(want))

	for _, name := range want {
		for _, module := range have {
			if module.Name == name {
				loaded = append(loaded, module.Name)
			}
		}
	}

	return loaded
}
