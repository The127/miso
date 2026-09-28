package tree

import (
	"testing"

	"golang.org/x/sys/unix"
)

// RaceOnce has the next lookup of a source in its stage answer EAGAIN, as
// the kernel does when a rename somewhere raced it, for the rest of a test.
func RaceOnce(t *testing.T) {
	t.Helper()

	was := openat2
	raced := false
	openat2 = func(dirfd int, path string, how *unix.OpenHow) (int, error) {
		if !raced {
			raced = true

			return -1, unix.EAGAIN
		}

		return was(dirfd, path, how)
	}

	t.Cleanup(func() { openat2 = was })
}
