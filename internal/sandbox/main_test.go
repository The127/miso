//go:build vmtest

package sandbox_test

import (
	"testing"

	"github.com/The127/miso/internal/sandbox"
	"github.com/The127/miso/internal/vmtest"
)

func TestMain(m *testing.M) {
	// the helper of a run is this binary too, and it must not run the tests
	sandbox.Helper()
	vmtest.Main(m)
}
