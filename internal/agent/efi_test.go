package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/The127/miso/internal/agent"
)

func TestAnArm64ImageBootsThroughItsArm64SystemdBootAndStub(t *testing.T) {
	// act
	efi := agent.EFIOf("arm64")

	// assert
	assert.Equal(t, agent.EFI{
		Loader:   "usr/lib/systemd/boot/efi/systemd-bootaa64.efi",
		Stub:     "usr/lib/systemd/boot/efi/linuxaa64.efi.stub",
		Fallback: "BOOTAA64.EFI",
	}, efi)
}
