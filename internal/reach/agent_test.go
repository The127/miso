package reach_test

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/reach"
	"github.com/The127/miso/internal/vport"
)

// onPort is a VM on a host without vsock, whose agent is reached over the
// host's end of its virtio port.
type onPort struct {
	port *os.File
}

func (onPort) CID() uint32 { return 0 }

func (v onPort) Port() *os.File { return v.port }

// port is a connected pair standing in for a virtio port: the host's end
// and the VM's.
func port(t *testing.T) (*os.File, *os.File) {
	t.Helper()

	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK, 0)
	require.NoError(t, err)
	host, machine := os.NewFile(uintptr(fds[0]), "host"), os.NewFile(uintptr(fds[1]), "machine")
	t.Cleanup(func() {
		_ = host.Close()
		_ = machine.Close()
	})

	return host, machine
}

func TestTheAgentOfAVMWithAVirtioPortIsReachedOverIt(t *testing.T) {
	// arrange
	host, machine := port(t)
	agent, err := vport.Listen(machine)
	require.NoError(t, err)
	dial, err := reach.Agent(onPort{host})
	require.NoError(t, err)

	// act
	dialed := make(chan error, 1)
	go func() {
		conn, err := dial()
		if err == nil {
			_, err = conn.Write([]byte("hello"))
		}

		// a failed dial must not leave the agent waiting
		if err != nil {
			_ = machine.Close()
		}

		dialed <- err
	}()
	accepted, accepting := agent.Accept()
	require.NoError(t, <-dialed)
	require.NoError(t, accepting)
	got := make([]byte, 5)
	_, err = io.ReadFull(accepted, got)

	// assert
	require.NoError(t, err)
	assert.Equal(t, "hello", string(got))
}
