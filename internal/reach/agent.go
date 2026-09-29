package reach

import (
	"errors"
	"io"
	"os"

	"github.com/The127/miso/internal/vport"
	"github.com/The127/miso/internal/vsock"
)

// ErrNoSockets is a VM on vsock without the sockets of its namespace. The
// host's own vsock would share the VM with everything else on it.
var ErrNoSockets = errors.New("a VM on vsock is reached only through the sockets of its namespace")

// VM is a builder VM whose agent the host connects to.
type VM interface {
	// CID is where the host reaches the VM over vsock.
	CID() uint32

	// Port is the host's end of the VM's virtio port, nil on a host with
	// vsock.
	Port() *os.File
}

// Agent is how the host connects to the agent of the VM, once for each
// request. Socket makes the sockets the host dials on, in the vsock
// namespace the VM runs in.
func Agent(vm VM, socket func() (*os.File, error)) (func() (io.ReadWriteCloser, error), error) {
	if port := vm.Port(); port != nil {
		dialer, err := vport.Connect(port)
		if err != nil {
			return nil, err
		}

		return dialer.Dial, nil
	}

	if socket == nil {
		return nil, ErrNoSockets
	}

	return func() (io.ReadWriteCloser, error) { return vsock.DialOn(socket, vm.CID(), vsock.AgentPort) }, nil
}
