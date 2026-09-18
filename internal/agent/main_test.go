//go:build vmtest

package agent_test

import (
	"testing"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/vmtest"
)

func TestMain(m *testing.M) {
	// the helper of a run is this binary too, and it must not run the tests
	agent.Helper()
	vmtest.Main(m)
}
