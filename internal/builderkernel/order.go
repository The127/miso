package builderkernel

// order is the order to load the wanted modules in, each of them after the
// modules it depends on.
func order(have []info, want ...string) []string {
	modules := make(map[string]info, len(have))
	for _, module := range have {
		modules[module.Name] = module
	}

	done := make(map[string]bool, len(modules))

	loaded := make([]string, 0, len(want))
	for _, name := range want {
		loaded = load(modules, done, loaded, name)
	}

	return loaded
}

// load appends the module and what it depends on, the dependencies first.
func load(modules map[string]info, done map[string]bool, loaded []string, name string) []string {
	module, found := modules[name]
	if !found || done[name] {
		return loaded
	}

	done[name] = true

	for _, dependency := range module.Depends {
		loaded = load(modules, done, loaded, dependency)
	}

	return append(loaded, name)
}
