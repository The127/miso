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
// image of arch. A TARGETARCH of the build file replaces miso's.
func envOf(under rootfs, arch string) []string {
	if slices.ContainsFunc(under.env, func(set string) bool { return strings.HasPrefix(set, "TARGETARCH=") }) {
		return slices.Clone(under.env)
	}

	return append([]string{"TARGETARCH=" + arch}, under.env...)
}
