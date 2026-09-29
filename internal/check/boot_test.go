//go:build kvm

package check_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/kvmtest"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

func TestChecksRunInABootedImageAndGiveTheirExitCodesWithKVM(t *testing.T) {
	// arrange
	boot := check.Boot{
		Driver:    qemu.Driver{Binary: "qemu-system-x86_64"},
		Namespace: vsocknstest.Private(t),
		Firmware:  kvmtest.Firmware(t),
		Image:     kvmtest.Image(t, "debian:sid"),
		Dir:       t.TempDir(),
		Console:   io.Discard,
		Patience:  time.Minute,
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
		Driver:    qemu.Driver{Binary: "qemu-system-x86_64"},
		Namespace: vsocknstest.Private(t),
		Firmware:  kvmtest.Firmware(t),
		Image:     qemu.Disk{Path: "/nonexistent/image.qcow2", Format: "qcow2", Serial: "image", Access: qemu.Snapshot},
		Dir:       t.TempDir(),
		Console:   io.Discard,
		Patience:  time.Minute,
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

func TestChecksGivenUpBeforeTheBootSayTheyWereGivenUpWithKVM(t *testing.T) {
	// arrange
	boot := check.Boot{
		Driver:    qemu.Driver{Binary: "qemu-system-x86_64"},
		Namespace: vsocknstest.Private(t),
		Firmware:  kvmtest.Firmware(t),
		Image:     kvmtest.Image(t, "debian:sid"),
		Dir:       t.TempDir(),
		Console:   io.Discard,
		Patience:  time.Minute,
	}
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
	defer cancel()

	// act
	_, err := boot.Run(ctx, []string{"true"})

	// assert
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestAnImageTheFirmwareCannotBootFailsItsChecksOnceThePatienceIsUpWithKVM(t *testing.T) {
	// arrange
	blank := filepath.Join(t.TempDir(), "blank.raw")
	require.NoError(t, os.WriteFile(blank, make([]byte, 1<<20), 0o600))
	boot := check.Boot{
		Driver:    qemu.Driver{Binary: "qemu-system-x86_64"},
		Namespace: vsocknstest.Private(t),
		Firmware:  kvmtest.Firmware(t),
		Image:     qemu.Disk{Path: blank, Format: "raw", Serial: "image", Access: qemu.Snapshot},
		Dir:       t.TempDir(),
		Console:   io.Discard,
		Patience:  5 * time.Second,
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
		assert.ErrorContains(t, err, "the image did not boot within 5s")
	case <-time.After(30 * time.Second):
		assert.Fail(t, "the checks of an image the firmware cannot boot never ended")
	}
}
