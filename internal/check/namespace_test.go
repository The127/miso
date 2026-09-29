package check_test

import (
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsockns"
	"github.com/The127/miso/internal/vsockns/vsocknstest"
)

func TestABootWithoutAVsockNamespaceIsRefused(t *testing.T) {
	// arrange
	boot := check.Boot{Dir: t.TempDir(), Patience: time.Second}

	// act
	_, err := boot.Run(t.Context(), []string{"true"})

	// assert
	assert.ErrorIs(t, err, check.ErrNoNamespace)
}

// errNoDevice is why a namespace without a vsock device hands over none.
var errNoDevice = errors.New("no vsock device here")

// withoutDevice is a namespace whose host has no vsock device.
type withoutDevice struct {
	*vsockns.Namespace
}

func (withoutDevice) Device() (*os.File, error) {
	return nil, errNoDevice
}

func TestABootWithoutAVsockDeviceIsRefusedWithWhy(t *testing.T) {
	// arrange
	boot := check.Boot{
		Driver:    lingering(t),
		Namespace: withoutDevice{vsocknstest.Private(t)},
		Image:     qemu.Disk{Path: "/o/image.raw", Format: "raw", Serial: "image", Access: qemu.Snapshot},
		Dir:       t.TempDir(),
		Console:   io.Discard,
		Patience:  time.Minute,
	}

	// act
	_, err := boot.Run(t.Context(), []string{"true"})

	// assert
	assert.ErrorIs(t, err, errNoDevice)
}
