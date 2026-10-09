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

// envOf is the environment of a run or a shell on a root file system for an
// image of arch. TARGETARCH comes first, so that a variable of the build file
// of the same name wins.
func envOf(under rootfs, arch string) []string {
	return append([]string{"TARGETARCH=" + arch}, under.env...)
}
