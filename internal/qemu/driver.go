package qemu

import (
	"context"
	"os"
	"os/exec"
	"slices"
)

// vhostVsock is the host's device on which a machine's CID is claimed.
const vhostVsock = "/dev/vhost-vsock"

// deviceFD is where QEMU finds the vsock device or its end of the agent's
// port, the first of a process's extra files after standard input, output
// and error.
const deviceFD = 3

// Driver starts builder VMs with the QEMU binary it names.
type Driver struct {
	Binary string

	// OpenVsock opens the vsock device the machine runs on, the host's own
	// when nil. A VM is in the network namespace its device was opened in.
	OpenVsock func() (*os.File, error)

	// WithoutKVM hears why the machine runs on TCG. It may be nil.
	WithoutKVM func(why error)

	// WithoutVsock hears why the machine's agent is reached over a virtio
	// port. It may be nil.
	WithoutVsock func(why error)
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

	reach, err := reachFor(d.openVsock)
	if err != nil {
		return nil, err
	}

	if reach.withoutVsock != nil {
		d.withoutVsock(reach.withoutVsock)
	}

	// QEMU holds its own copy of its end, and of a vsock device the CID with
	// it, so the host's copy goes once QEMU runs
	defer func() { _ = reach.machine.Close() }()

	if err := inheritNothing(); err != nil {
		return nil, err
	}

	//nolint:gosec // running the QEMU the caller names with the machine it describes is the job
	command := exec.CommandContext(ctx, d.Binary, slices.Concat(args, accelerated, reach.args)...)
	command.ExtraFiles = []*os.File{reach.machine}
	command.Stdout = machine.Console

	if machine.Temp != "" {
		command.Env = append(os.Environ(), "TMPDIR="+machine.Temp)
	}

	vm, err := run(command, reach.cid, reach.host)
	if err != nil {
		_ = reach.close()

		return nil, err
	}

	vm.withoutKVM = why != nil

	return vm, nil
}

// openVsock opens the device the machine runs on.
func (d Driver) openVsock() (*os.File, error) {
	if d.OpenVsock != nil {
		return d.OpenVsock()
	}

	return openVsock(vhostVsock)
}

// withoutKVM is only a notice, the machine still runs, only slower.
func (d Driver) withoutKVM(why error) {
	if d.WithoutKVM != nil {
		d.WithoutKVM(why)
	}
}

// withoutVsock is only a notice, the agent is still reached over the port.
func (d Driver) withoutVsock(why error) {
	if d.WithoutVsock != nil {
		d.WithoutVsock(why)
	}
}
