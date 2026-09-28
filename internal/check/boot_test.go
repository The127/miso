//go:build kvm

package check_test

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/kvmtest"
	"github.com/The127/miso/internal/qemu"
)

func TestChecksRunInABootedImageAndGiveTheirExitCodesWithKVM(t *testing.T) {
	// arrange
	boot := check.Boot{
		Driver:   qemu.Driver{Binary: "qemu-system-x86_64"},
		Firmware: kvmtest.Firmware(t),
		Image:    kvmtest.Image(t, "debian:sid"),
		Dir:      t.TempDir(),
		Console:  io.Discard,
	}

	// act
	results, err := boot.Run(t.Context(), []string{"command -v sh", "false"})

	// assert
	require.NoError(t, err)
	assert.Equal(t, []check.Result{{Output: "/usr/bin/sh\n", Code: 0}, {Output: "", Code: 1}}, results)
}
