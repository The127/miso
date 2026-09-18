//go:build vmtest

package vmtest_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/vmtest"
)

func TestMain(m *testing.M) {
	vmtest.Main(m)
}

func TestTheTestsRunAsRoot(t *testing.T) {
	// act
	uid := os.Getuid()

	// assert
	assert.Equal(t, 0, uid)
}
