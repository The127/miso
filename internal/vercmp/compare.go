package vercmp

import (
	"cmp"
	"strings"
)

// Compare is negative when version a is older than b, zero when both are
// the same version and positive when a is newer.
func Compare(a, b string) int {
	// a longer number is a bigger one, however many digits it has
	if len(a) != len(b) {
		return cmp.Compare(len(a), len(b))
	}

	return strings.Compare(a, b)
}
