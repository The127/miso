package check_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/qemu"
)

func TestTheImageListensOnVsockForChecksOnceItLeftTheInitrd(t *testing.T) {
	// act
	credentials := check.Credentials(12345)

	// assert
	assert.Contains(t, credentials, qemu.Credential{Name: "systemd.extra-unit.miso-check.socket", Value: []byte(`[Unit]
Description=miso check shell
ConditionPathExists=!/etc/initrd-release

[Socket]
ListenStream=vsock::5000
Accept=yes
`)})
}
