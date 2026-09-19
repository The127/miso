//go:build vmtest

package agent_test

import (
	"os"
	"testing"

	"github.com/The127/miso/internal/sandbox"
	"github.com/The127/miso/internal/vmtest"
)

func TestMain(m *testing.M) {
	// a bare layer runs this binary as its shell, which does nothing
	if os.Args[0] == "/bin/sh" {
		os.Exit(0)
	}

	// the helper of a run is this binary too, and it must not run the tests
	sandbox.Helper()
	vmtest.Main(m)
}
