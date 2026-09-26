package qemu

import (
	"context"
	"math/rand/v2"
	"os/exec"
)

// vhostVsock is the host's device on which a machine's CID is claimed.
const vhostVsock = "/dev/vhost-vsock"

// Driver starts builder VMs with the QEMU binary it names.
type Driver struct {
	Binary string
}

// Start boots the machine.
func (d Driver) Start(_ context.Context, machine Machine) (*VM, error) {
	args, err := arguments(machine)
	if err != nil {
		return nil, err
	}

	device, err := openVsock(vhostVsock)
	if err != nil {
		return nil, err
	}

	defer func() { _ = device.Close() }()

	cid, err := claim(func(cid uint32) error { return takeCID(device, cid) }, rand.Uint32)
	if err != nil {
		return nil, err
	}

	//nolint:gosec // running the QEMU the caller names with the machine it describes is the job
	command := exec.Command(d.Binary, append(args, vsock(cid, 3)...)...)
	if err := command.Start(); err != nil {
		return nil, err
	}

	vm := &VM{cid: cid, done: make(chan struct{})}

	go func() {
		_ = command.Wait()

		close(vm.done)
	}()

	return vm, nil
}
