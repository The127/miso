package build

import (
	"slices"
	"strings"

	"github.com/The127/miso/internal/imagefile"
)

// withVariable keeps a variable set twice in one place, so that nothing
// that reads the first match gets the old value.
func withVariable(env []string, variable imagefile.Env) []string {
	set := slices.Clone(env)
	assignment := variable.Key + "=" + variable.Value
	at := slices.IndexFunc(set, func(earlier string) bool { return strings.HasPrefix(earlier, variable.Key+"=") })
	if at < 0 {
		return append(set, assignment)
	}

	set[at] = assignment

	return set
}
