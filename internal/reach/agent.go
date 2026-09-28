package reach

import (
	"io"

	"github.com/The127/miso/internal/vsock"
)

// VM is a builder VM whose agent the host connects to.
type VM interface {
	// CID is where the host reaches the VM over vsock.
	CID() uint32
}

// Agent is how the host connects to the agent of the VM, once for each
// request.
func Agent(vm VM) func() (io.ReadWriteCloser, error) {
	return func() (io.ReadWriteCloser, error) { return vsock.Dial(vm.CID(), vsock.AgentPort) }
}
