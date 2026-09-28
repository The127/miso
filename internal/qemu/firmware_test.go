package qemu_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/qemu"
)

func TestAMachineWithFirmwareBootsThroughItInsteadOfAKernel(t *testing.T) {
	// arrange
	machine := qemu.Machine{Boot: qemu.Firmware{Code: "/f/OVMF_CODE.fd", Vars: "/f/vars.fd"}}

	// act
	args, err := qemu.Arguments(machine)

	// assert
	require.NoError(t, err)
	assert.Equal(t, []string{
		"if=pflash,format=raw,readonly=on,file=/f/OVMF_CODE.fd",
		"if=pflash,format=raw,file=/f/vars.fd",
	}, valuesOf(args, "-drive"))
	assert.NotContains(t, args, "-kernel")
}
