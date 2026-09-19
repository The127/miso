package builderkernel

import (
	"fmt"
	"strings"
)

// order is the order to load the wanted modules in, each of them after the
// modules it depends on. What the kernel has built in is no module and is
// left out.
func order(have []info, builtin []string, want ...string) ([]string, error) {
	modules := make(map[string]info, len(have))
	for _, module := range have {
		modules[named(module.Name)] = module
	}

	done := make(map[string]bool, len(modules)+len(builtin))
	for _, name := range builtin {
		done[named(name)] = true
	}

	loaded := make([]string, 0, len(want))
	for _, name := range want {
		var err error

		loaded, err = load(modules, done, loaded, name)
		if err != nil {
			return nil, err
		}
	}

	return loaded, nil
}

// named is a module's name as the kernel reads it, which takes a dash and an
// underscore for the same character.
func named(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// load appends the module and what it depends on, the dependencies first.
func load(modules map[string]info, done map[string]bool, loaded []string, name string) ([]string, error) {
	key := named(name)
	if done[key] {
		return loaded, nil
	}

	module, found := modules[key]
	if !found {
		return nil, fmt.Errorf("the package holds no module %s", name)
	}

	done[key] = true

	for _, dependency := range module.Depends {
		var err error

		loaded, err = load(modules, done, loaded, dependency)
		if err != nil {
			return nil, err
		}
	}

	return append(loaded, module.Name), nil
}
