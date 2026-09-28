package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineHandsItsCredentialsToSystemdThroughSMBIOS(t *testing.T) {
	// arrange
	machine := qemu.Machine{Credentials: []qemu.Credential{
		{Name: "vmm.notify_socket", Value: []byte("vsock-stream:2:12345")},
		{Name: "systemd.unit-dropin.sockets.target", Value: []byte("[Unit]\nWants=miso-check.socket\n")},
	}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{
		"type=11,value=io.systemd.credential.binary:vmm.notify_socket=dnNvY2stc3RyZWFtOjI6MTIzNDU=",
		"type=11,value=io.systemd.credential.binary:systemd.unit-dropin.sockets.target=W1VuaXRdCldhbnRzPW1pc28tY2hlY2suc29ja2V0Cg==",
	}, valuesOf(args, "-smbios"))
}
