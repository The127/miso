package check_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/qemu"
)

// lingering is a QEMU whose console is written to a moment after it was
// killed, by a child that still holds it.
func lingering(t *testing.T) qemu.Driver {
	t.Helper()

	binary := filepath.Join(t.TempDir(), "qemu")
	//nolint:gosec // the fake QEMU must be executable
	require.NoError(t, os.WriteFile(binary, []byte("#!/bin/sh\n(sleep 0.3; echo late) &\nexec sleep 60\n"), 0o700))

	return qemu.Driver{Binary: binary}
}

func TestChecksReturnOnlyOnceTheirQEMUIsGone(t *testing.T) {
	// arrange
	var console bytes.Buffer
	boot := check.Boot{
		Driver:    lingering(t),
		Namespace: private(t),
		Image:     qemu.Disk{Path: "/o/image.raw", Format: "raw", Serial: "image", Access: qemu.Snapshot},
		Dir:       t.TempDir(),
		Console:   &console,
		Patience:  100 * time.Millisecond,
	}

	// act
	_, err := boot.Run(t.Context(), []string{"true"})

	// assert
	require.ErrorContains(t, err, "did not boot within")
	assert.Contains(t, console.String(), "late")
}
