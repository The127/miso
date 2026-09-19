package cachedisk

import "testing"

// SearchAlso has Make look for mkfs in dirs instead of the sbin
// directories, for the rest of a test.
func SearchAlso(t *testing.T, dirs ...string) {
	t.Helper()

	was := searchAlso
	searchAlso = dirs

	t.Cleanup(func() { searchAlso = was })
}
