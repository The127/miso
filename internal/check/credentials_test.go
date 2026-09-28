package check_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/qemu"
)

func TestTheImageTellsTheHostOverVsockWhenItHasBooted(t *testing.T) {
	// act
	credentials := check.Credentials(12345)

	// assert
	assert.Contains(t, credentials, qemu.Credential{Name: "vmm.notify_socket", Value: []byte("vsock-stream:2:12345")})
}
