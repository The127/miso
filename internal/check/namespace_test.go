//go:build kvm

package check_test

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/kvmtest"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsockns"
)

func TestChecksRunInAnImageBootedInAVsockNamespaceWithKVM(t *testing.T) {
	// arrange
	namespace, err := vsockns.Open()
	if errors.Is(err, vsockns.ErrNotPrivate) {
		t.Skip("this host cannot keep vsock private")
	}

	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, namespace.Close()) })
	boot := check.Boot{
		Driver:    qemu.Driver{Binary: "qemu-system-x86_64"},
		Namespace: namespace,
		Firmware:  kvmtest.Firmware(t),
		Image:     kvmtest.Image(t, "debian:sid"),
		Dir:       t.TempDir(),
		Console:   io.Discard,
		Patience:  time.Minute,
	}

	// act
	results, err := boot.Run(t.Context(), []string{"command -v sh"})

	// assert
	require.NoError(t, err)
	assert.Equal(t, []check.Result{{Output: "/usr/bin/sh\n", Code: 0}}, results)
}
