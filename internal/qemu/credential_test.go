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

func TestAMicrovmHandsItsCredentialsToSystemdOnTheKernelCommandLine(t *testing.T) {
	// arrange
	machine := qemu.Machine{
		Microvm: true,
		Boot:    qemu.Kernel{Image: "/vmlinux", Initramfs: "/initrd.img", CommandLine: "root=/dev/vda rw"},
		Credentials: []qemu.Credential{
			{Name: "vmm.notify_socket", Value: []byte("vsock-stream:2:12345")},
			{Name: "systemd.unit-dropin.sockets.target", Value: []byte("[Unit]\nWants=miso-check.socket\n")},
		},
	}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "root=/dev/vda rw"+
		" systemd.set_credential_binary=vmm.notify_socket:dnNvY2stc3RyZWFtOjI6MTIzNDU="+
		" systemd.set_credential_binary=systemd.unit-dropin.sockets.target:W1VuaXRdCldhbnRzPW1pc28tY2hlY2suc29ja2V0Cg==",
		valueOf(t, args, "-append"))
	assert.NotContains(t, args, "-smbios")
}

func TestAnArm64KernelGetsItsCredentialsOnTheCommandLine(t *testing.T) {
	// arrange
	machine := qemu.Machine{
		Arch:        "arm64",
		Boot:        qemu.Kernel{Image: "/Image", Initramfs: "/initrd.img", CommandLine: "root=/dev/vda rw"},
		Credentials: []qemu.Credential{{Name: "vmm.notify_socket", Value: []byte("vsock-stream:2:12345")}},
	}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "root=/dev/vda rw systemd.set_credential_binary=vmm.notify_socket:dnNvY2stc3RyZWFtOjI6MTIzNDU=", valueOf(t, args, "-append"))
	assert.NotContains(t, args, "-smbios")
}

func TestAnArm64FirmwareGetsItsCredentialsThroughSMBIOS(t *testing.T) {
	// arrange
	machine := qemu.Machine{
		Arch:        "arm64",
		Boot:        qemu.Firmware{Code: "code.fd", Vars: "vars.fd"},
		Credentials: []qemu.Credential{{Name: "vmm.notify_socket", Value: []byte("vsock-stream:2:12345")}},
	}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{"type=11,value=io.systemd.credential.binary:vmm.notify_socket=dnNvY2stc3RyZWFtOjI6MTIzNDU="}, valuesOf(args, "-smbios"))
	assert.NotContains(t, args, "-kernel")
}

func TestAMicrovmThatDoesNotBootAKernelIsRefused(t *testing.T) {
	// arrange
	machine := qemu.Machine{Microvm: true, Boot: qemu.Firmware{Code: "code.fd", Vars: "vars.fd"}}

	// act
	_, err := qemu.Arguments(machine)

	// assert
	assert.ErrorContains(t, err, "needs a kernel")
}

func TestAMicrovmWithoutACommandLineHasNoSpaceBeforeItsCredentials(t *testing.T) {
	// arrange
	machine := qemu.Machine{
		Microvm:     true,
		Boot:        qemu.Kernel{Image: "/vmlinux", Initramfs: "/initrd.img"},
		Credentials: []qemu.Credential{{Name: "vmm.notify_socket", Value: []byte("vsock-stream:2:12345")}},
	}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "systemd.set_credential_binary=vmm.notify_socket:dnNvY2stc3RyZWFtOjI6MTIzNDU=", valueOf(t, args, "-append"))
}

func TestAMicrovmWhoseCredentialsDoNotFitTheKernelCommandLineIsRefused(t *testing.T) {
	// arrange
	machine := qemu.Machine{
		Microvm:     true,
		Boot:        qemu.Kernel{Image: "/vmlinux", Initramfs: "/initrd.img", CommandLine: "root=/dev/vda"},
		Credentials: []qemu.Credential{{Name: "systemd.extra-unit.big.service", Value: make([]byte, 2000)}},
	}

	// act
	_, err := qemu.Arguments(machine)

	// assert
	assert.ErrorContains(t, err, "2047")
}

func TestAMicrovmWithACredentialNameThatWouldSplitTheCommandLineIsRefused(t *testing.T) {
	// arrange
	machine := qemu.Machine{
		Microvm:     true,
		Boot:        qemu.Kernel{Image: "/vmlinux", Initramfs: "/initrd.img"},
		Credentials: []qemu.Credential{{Name: "name with space", Value: []byte("x")}},
	}

	// act
	_, err := qemu.Arguments(machine)

	// assert
	assert.ErrorContains(t, err, "name with space")
}

func TestAMicrovmWithoutAnyBootIsRefused(t *testing.T) {
	// arrange
	machine := qemu.Machine{Microvm: true}

	// act
	_, err := qemu.Arguments(machine)

	// assert
	assert.ErrorContains(t, err, "needs a kernel")
}
