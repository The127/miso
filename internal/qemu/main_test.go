package qemu_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestMain stands in for QEMU when a test starts this binary as one, and
// writes the arguments it was given to the file MISO_FAKE_QEMU names, and a
// line to its console. With MISO_FAKE_QEMU_FAIL it says that on its standard
// error and fails, the way QEMU refuses what it cannot run. With
// MISO_FAKE_QEMU_HANG it then hangs like a running machine, for a while
// only, so a failed test leaves nothing behind for long.
func TestMain(m *testing.M) {
	if recorded := os.Getenv("MISO_FAKE_QEMU"); recorded != "" {
		//nolint:gosec // the test that started this binary names the file
		if err := os.WriteFile(recorded, []byte(strings.Join(os.Args[1:], "\n")), 0o600); err != nil {
			os.Exit(2)
		}

		fmt.Println("fake QEMU console")

		if refusal := os.Getenv("MISO_FAKE_QEMU_FAIL"); refusal != "" {
			fmt.Fprintln(os.Stderr, refusal)
			os.Exit(1)
		}

		if os.Getenv("MISO_FAKE_QEMU_HANG") != "" {
			time.Sleep(30 * time.Second)
		}

		return
	}

	os.Exit(m.Run())
}
