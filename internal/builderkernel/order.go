package builderkernel

import (
	"fmt"
	"slices"
	"strings"
)

// order is the order to load the wanted modules in, each of them after the
// modules it depends on. What the kernel has built in is no module and is
// left out.
func order(have []info, builtin []string, want ...string) ([]string, error) {
	l := loading(have, builtin)

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

// loading is a loader over the modules there are, with what the kernel has
// built in already behind it.
func loading(have []info, builtin []string) loader {
	l := loader{
		modules: make(map[string]info, len(have)),
		done:    make(map[string]bool, len(have)+len(builtin)),
		loaded:  make([]string, 0, len(have)),
	}

	for _, module := range have {
		l.modules[named(module.Name)] = module
	}

	for _, name := range builtin {
		l.done[named(name)] = true
	}

	return l
}

// loader walks the modules it is given, collecting them in the order to load.
// walking is the path it is on, which is how it sees a module depend on one
// that depends on it.
type loader struct {
	modules map[string]info
	done    map[string]bool
	walking []string
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

	if at := slices.Index(l.walking, key); at >= 0 {
		cycle := append(slices.Clone(l.walking[at:]), key)

		return fmt.Errorf("the modules depend on each other: %s", strings.Join(cycle, ", "))
	}

	l.walking = append(l.walking, key)

	for _, dependency := range module.Depends {
		if err := l.load(dependency); err != nil {
			return err
		}
	}

	l.walking = l.walking[:len(l.walking)-1]
	l.done[key] = true
	l.loaded = append(l.loaded, module.Name)

	return nil
}
