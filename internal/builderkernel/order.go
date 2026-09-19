package builderkernel

// order is the order to load the wanted modules in, each of them after the
// modules it depends on.
func order(have []info, want ...string) []string {
	modules := make(map[string]info, len(have))
	for _, module := range have {
		modules[module.Name] = module
	}

	loaded := make([]string, 0, len(want))
	for _, name := range want {
		loaded = load(modules, loaded, name)
	}

	return loaded
}

// load appends the module and what it depends on, the dependencies first.
func load(modules map[string]info, loaded []string, name string) []string {
	module, found := modules[name]
	if !found {
		return loaded
	}

	for _, dependency := range module.Depends {
		loaded = load(modules, loaded, dependency)
	}

	return append(loaded, name)
}
