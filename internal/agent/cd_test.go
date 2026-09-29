//go:build vmtest

package agent_test

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/The127/miso/internal/agent"
	"github.com/The127/miso/internal/protocol"
)

// stubbed adds to the base the systemd-boot an image brings and a UKI stub
// that says it is the version, as systemd's own stub does, and answers the
// layers of that image.
func stubbed(t *testing.T, worker *agent.Agent, version string) []string {
	t.Helper()

	return withStub(t, worker, "MZ #### LoaderInfo: systemd-stub "+version+" ####\n")
}

// withStub adds to the base the systemd-boot an image brings and a UKI
// stub of the bytes, and answers the layers of that image.
func withStub(t *testing.T, worker *agent.Agent, stub string) []string {
	t.Helper()

	boot := `mkdir -p /usr/lib/systemd/boot/efi && cd /usr/lib/systemd/boot/efi && echo loader > systemd-bootx64.efi && printf '%s' "$STUB" > linuxx64.efi.stub`
	run := protocol.Run{Key: "boot", Layers: []string{"base"}, Command: boot, Env: []string{"STUB=" + stub}}
	code, err := worker.Run(context.Background(), run, io.Discard)
	require.NoError(t, err)
	require.Equal(t, 0, code)

	return []string{"base", "boot"}
}

func TestAnISOOfAnImageWhoseStubIsOlderThan261FailsNamingItsVersion(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	disk := protocol.Disk{Key: "disk", Layers: stubbed(t, worker, "257.9-1"), Tools: fakeTools(t, worker, writesDisk), ElTorito: true}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.Error(t, err)
	assert.ErrorContains(t, err, `"257.9-1"`)
	assert.ErrorContains(t, err, "261")
}

func TestAnISOOfAnImageWhoseStubDoesNotSayItsVersionFails(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	disk := protocol.Disk{Key: "disk", Layers: withStub(t, worker, "MZ stub\n"), Tools: fakeTools(t, worker, writesDisk), ElTorito: true}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "261")
}

func TestAnISOOfAnImageWhoseStubBreaksOffItsVersionFails(t *testing.T) {
	// arrange
	worker := mountedBase(t, t.TempDir())
	disk := protocol.Disk{Key: "disk", Layers: withStub(t, worker, "MZ #### LoaderInfo: systemd-stub 262"), Tools: fakeTools(t, worker, writesDisk), ElTorito: true}

	// act
	err := worker.Disk(context.Background(), disk, io.Discard)

	// assert
	require.Error(t, err)
	assert.ErrorContains(t, err, "261")
}
