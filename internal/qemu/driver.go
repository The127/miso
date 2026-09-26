package qemu

import (
	"bytes"
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// vhostVsock is the host's device on which a machine's CID is claimed.
const vhostVsock = "/dev/vhost-vsock"

// deviceFD is where QEMU finds the device, the first of a process's extra
// files after standard input, output and error.
const deviceFD = 3

// Driver starts builder VMs with the QEMU binary it names.
type Driver struct {
	Binary string
}

// Start boots the machine, which is killed when ctx is done. Killing is
// safe, the cache disk is written to survive a crash of the machine.
func (d Driver) Start(ctx context.Context, machine Machine) (*VM, error) {
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
	command := exec.CommandContext(ctx, d.Binary, append(args, vsock(cid, deviceFD)...)...)
	// QEMU keeps the device open, and with it the CID, which the host lets
	// go once QEMU runs
	command.ExtraFiles = []*os.File{device}
	command.Stdout = machine.Console
	// a QEMU left behind by a miso that was killed would hold its CID and
	// its cache disk for good
	command.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}

	// QEMU says on its standard error why it cannot run a machine
	var refusal bytes.Buffer
	command.Stderr = &refusal
	if err := command.Start(); err != nil {
		return nil, err
	}

	vm := &VM{cid: cid, done: make(chan struct{})}

	go func() {
		if err := command.Wait(); err != nil {
			vm.err = fmt.Errorf("QEMU stopped: %w: %s", err, strings.TrimSpace(refusal.String()))
		}

		close(vm.done)
	}()

	return vm, nil
}
