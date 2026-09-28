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
// request.
func Agent(vm VM) (func() (io.ReadWriteCloser, error), error) {
	if port := vm.Port(); port != nil {
		dialer, err := vport.Connect(port)
		if err != nil {
			return nil, err
		}

		return dialer.Dial, nil
	}

	return func() (io.ReadWriteCloser, error) { return vsock.Dial(vm.CID(), vsock.AgentPort) }, nil
}
