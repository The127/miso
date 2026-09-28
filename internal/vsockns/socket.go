package vsockns

import (
	"fmt"
	"os"
)

// Socket is a fresh vsock socket made inside the namespace. A socket stays
// in the namespace it was made in, whoever binds, listens or connects it
// later.
func (n *Namespace) Socket() (*os.File, error) {
	_, files, err := n.ask(socketQuestion)
	if err != nil {
		return nil, err
	}

	if len(files) != 1 {
		return nil, fmt.Errorf("the helper handed over %d sockets", len(files))
	}

	return os.NewFile(uintptr(files[0]), "vsock"), nil
}
