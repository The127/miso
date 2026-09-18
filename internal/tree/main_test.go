//go:build vmtest

package tree_test

import (
	"testing"

	"github.com/The127/miso/internal/vmtest"
)

func TestMain(m *testing.M) {
	vmtest.Main(m)
}
