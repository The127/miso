package qemu_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/qemu"
)

func TestAHostWithoutTheVsockDeviceIsToldWhichModuleItNeeds(t *testing.T) {
	// arrange
	missing := filepath.Join(t.TempDir(), "vhost-vsock")

	// act
	_, err := qemu.OpenVsock(missing)

	// assert
	assert.ErrorContains(t, err, missing)
	assert.ErrorContains(t, err, "vhost_vsock")
}
