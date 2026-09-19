package builderkernel

import (
	"fmt"
	"strings"
)

// order is the order to load the wanted modules in, each of them after the
// modules it depends on. What the kernel has built in is no module and is
// left out.
func order(have []info, builtin []string, want ...string) ([]string, error) {
	l := loader{
		modules: make(map[string]info, len(have)),
		done:    make(map[string]bool, len(have)+len(builtin)),
		loaded:  make([]string, 0, len(want)),
	}

	for _, module := range have {
		l.modules[named(module.Name)] = module
	}

	for _, name := range builtin {
		l.done[named(name)] = true
	}

	for _, name := range want {
		if err := l.load(name); err != nil {
			return nil, err
		}
	}

	return l.loaded, nil
}

// named is a module's name as the kernel reads it, which takes a dash and an
// underscore for the same character.
func named(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// loader walks the modules it is given, collecting them in the order to load.
type loader struct {
	modules map[string]info
	done    map[string]bool
	loaded  []string
}

// load adds the module and what it depends on, the dependencies first.
func (l *loader) load(name string) error {
	key := named(name)
	if l.done[key] {
		return nil
	}

	module, found := l.modules[key]
	if !found {
		return fmt.Errorf("the package holds no module %s", name)
	}

	l.done[key] = true

	for _, dependency := range module.Depends {
		if err := l.load(dependency); err != nil {
			return err
		}
	}

	l.loaded = append(l.loaded, module.Name)

	return nil
}
