package qemu

import (
	"context"
	"os"
	"os/exec"
	"slices"
)

// vhostVsock is the host's device on which a machine's CID is claimed.
const vhostVsock = "/dev/vhost-vsock"

// deviceFD is where QEMU finds the device, the first of a process's extra
// files after standard input, output and error.
const deviceFD = 3

// Driver starts builder VMs with the QEMU binary it names.
type Driver struct {
	Binary string

	// WithoutKVM hears why the machine runs on TCG. It may be nil.
	WithoutKVM func(why error)
}

// Start boots the machine, which is killed when ctx is done. Killing is
// safe, the cache disk is written to survive a crash of the machine.
func (d Driver) Start(ctx context.Context, machine Machine) (*VM, error) {
	return d.start(ctx, machine, kvmDevice)
}

func (d Driver) start(ctx context.Context, machine Machine, kvm string) (*VM, error) {
	args, err := arguments(machine)
	if err != nil {
		return nil, err
	}

	accelerated, why := accel(kvm)
	if why != nil {
		d.withoutKVM(why)
	}

	device, cid, err := holdCID(vhostVsock)
	if err != nil {
		return nil, err
	}

	defer func() { _ = device.Close() }()

	//nolint:gosec // running the QEMU the caller names with the machine it describes is the job
	command := exec.CommandContext(ctx, d.Binary, slices.Concat(args, accelerated, vsock(cid, deviceFD))...)
	// QEMU keeps the device open, and with it the CID, which the host lets
	// go once QEMU runs
	command.ExtraFiles = []*os.File{device}
	command.Stdout = machine.Console

	vm, err := run(command, cid)
	if err != nil {
		return nil, err
	}

	vm.withoutKVM = why != nil

	return vm, nil
}

// withoutKVM is only a notice, the machine still runs, only slower.
func (d Driver) withoutKVM(why error) {
	if d.WithoutKVM != nil {
		d.WithoutKVM(why)
	}
}
