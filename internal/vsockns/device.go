package vsockns

import (
	"fmt"
	"os"
)

// Device is the host's vhost-vsock device, opened inside the namespace. The
// device stays in the namespace it was opened in, and the VM that QEMU runs
// on it does too, wherever QEMU itself runs.
func (n *Namespace) Device() (*os.File, error) {
	_, files, err := n.ask(deviceQuestion)
	if err != nil {
		return nil, err
	}

	if len(files) != 1 {
		return nil, fmt.Errorf("the helper handed over %d devices", len(files))
	}

	return os.NewFile(uintptr(files[0]), "vhost-vsock"), nil
}
