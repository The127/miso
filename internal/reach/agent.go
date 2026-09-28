package reach

import (
	"io"
	"os"

	"github.com/The127/miso/internal/vport"
	"github.com/The127/miso/internal/vsock"
)

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
// namespace the VM runs in. The host's own vsock when nil.
func Agent(vm VM, socket func() (*os.File, error)) (func() (io.ReadWriteCloser, error), error) {
	if port := vm.Port(); port != nil {
		dialer, err := vport.Connect(port)
		if err != nil {
			return nil, err
		}

		return dialer.Dial, nil
	}

	if socket == nil {
		return func() (io.ReadWriteCloser, error) { return vsock.Dial(vm.CID(), vsock.AgentPort) }, nil
	}

	return func() (io.ReadWriteCloser, error) {
		dialing, err := socket()
		if err != nil {
			return nil, err
		}

		defer func() { _ = dialing.Close() }()

		return vsock.DialOn(dialing, vm.CID(), vsock.AgentPort)
	}, nil
}
