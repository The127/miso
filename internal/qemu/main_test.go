package qemu_test

import (
	"os"
	"strings"
	"testing"
)

// TestMain stands in for QEMU when a test starts this binary as one, and
// writes the arguments it was given to the file MISO_FAKE_QEMU names.
func TestMain(m *testing.M) {
	if recorded := os.Getenv("MISO_FAKE_QEMU"); recorded != "" {
		//nolint:gosec // the test that started this binary names the file
		if err := os.WriteFile(recorded, []byte(strings.Join(os.Args[1:], "\n")), 0o600); err != nil {
			os.Exit(2)
		}

		return
	}

	os.Exit(m.Run())
}
