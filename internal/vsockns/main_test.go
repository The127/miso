package vsockns_test

import (
	"os"
	"testing"

	"github.com/The127/miso/internal/vsockns"
)

func TestMain(m *testing.M) {
	// the helper of a namespace is this binary too, and it must not run the
	// tests
	vsockns.Helper()
	os.Exit(m.Run())
}
