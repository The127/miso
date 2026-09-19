//go:build vmtest

package boot_test

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTheRootOfTheTestsIsNotTheInitialRamfs(t *testing.T) {
	// act
	mounts, err := os.ReadFile("/proc/self/mountinfo")

	// assert
	require.NoError(t, err)
	var fstype string
	for line := range strings.Lines(string(mounts)) {
		fields, after, _ := strings.Cut(line, " - ")
		if strings.Fields(fields)[4] == "/" {
			fstype = strings.Fields(after)[0]
		}
	}

	require.NotEmpty(t, fstype, "no mount on /")
	// kernels before the empty mount under the initial ramfs refuse to
	// pivot_root out of it, and a run pivots
	assert.NotEqual(t, "rootfs", fstype)
}
