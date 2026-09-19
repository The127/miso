//go:build vmtest

package boot_test

import (
	"testing"

	"github.com/The127/miso/internal/vmtest"
)

func TestMain(m *testing.M) {
	vmtest.Main(m)
}
