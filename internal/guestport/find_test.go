package guestport_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/guestport"
)

// ports writes a sysfs with the virtio ports it names, each under its
// device name.
func ports(t *testing.T, named map[string]string) string {
	t.Helper()

	sys := t.TempDir()
	for device, name := range named {
		dir := filepath.Join(sys, "class", "virtio-ports", device)
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "name"), []byte(name+"\n"), 0o600))
	}

	return sys
}

func TestTheAgentsPortIsFoundByItsName(t *testing.T) {
	// arrange
	sys := ports(t, map[string]string{"vport0p1": "org.qemu.guest_agent.0", "vport1p1": "miso"})

	// act
	found, err := guestport.Find(sys, "/dev", "miso")

	// assert
	require.NoError(t, err)
	assert.Equal(t, "/dev/vport1p1", found)
}
