package check_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/qemu"
)

func TestACheckRunsInAShellThatEndsItsOutputWithTheExitCode(t *testing.T) {
	// act
	credentials := check.Credentials(12345)

	// assert
	assert.Contains(t, credentials, qemu.Credential{Name: "systemd.extra-unit.miso-check@.service", Value: []byte(`[Unit]
Description=miso check

[Service]
ExecStart=/bin/sh -c '/bin/sh; code=$$?; echo; echo "miso-exit $$code"'
StandardInput=socket
StandardOutput=socket
StandardError=socket
`)})
}
