//go:build kvm

package check_test

import (
	"io"
	"testing"
	"time"

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

func TestAnImageThatCannotBeBootedFailsItsChecksWithQEMUsReasonWithKVM(t *testing.T) {
	// arrange
	boot := check.Boot{
		Driver:   qemu.Driver{Binary: "qemu-system-x86_64"},
		Firmware: kvmtest.Firmware(t),
		Image:    qemu.Disk{Path: "/nonexistent/image.qcow2", Format: "qcow2", Serial: "image", Snapshot: true},
		Dir:      t.TempDir(),
		Console:  io.Discard,
	}
	failed := make(chan error, 1)

	// act
	go func() {
		_, err := boot.Run(t.Context(), []string{"true"})
		failed <- err
	}()

	// assert
	select {
	case err := <-failed:
		assert.ErrorContains(t, err, "the image did not boot")
		assert.ErrorContains(t, err, "/nonexistent/image.qcow2")
	case <-time.After(30 * time.Second):
		assert.Fail(t, "the checks of an image that cannot boot never ended")
	}
}
