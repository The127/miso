package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/reach"
	"github.com/The127/miso/internal/vsockns"
)

// bootFiles writes the builder kernel and an initramfs with this very miso
// as its init into a directory of their own.
func bootFiles(ctx context.Context, blobs *download.Store) (builder.Boot, error) {
	kernel, err := builderkernel.Ready(ctx, blobs)
	if err != nil {
		return builder.Boot{}, err
	}

	self, err := os.Executable()
	if err != nil {
		return builder.Boot{}, err
	}

	init, err := os.ReadFile(self)
	if err != nil {
		return builder.Boot{}, err
	}

	dir, err := os.MkdirTemp("", "miso-boot-")
	if err != nil {
		return builder.Boot{}, err
	}

	boot, err := builder.WriteBoot(dir, kernel, init)
	if err != nil {
		return builder.Boot{}, errors.Join(err, os.RemoveAll(dir))
	}

	return boot, nil
}

// inBuilder does the work with the builder VM booted, its console in a log
// file, and stops the VM afterwards. The work dials the VM's agent. What the
// user should know about how it runs goes to said.
func inBuilder(ctx context.Context, machine qemu.Machine, log string, said io.Writer, work func(vm *qemu.VM, dial func() (io.ReadWriteCloser, error)) error) error {
	console, err := os.Create(log)
	if err != nil {
		return err
	}

	defer func() { _ = console.Close() }()

	machine.Console = console

	running, stop := context.WithCancel(ctx)
	defer stop()

	driver := qemu.Driver{
		Binary: "qemu-system-x86_64",
		WithoutKVM: func(why error) {
			_, _ = fmt.Fprintln(said, "miso: without KVM the builder VM runs much slower:", why)
		},
		WithoutVsock: func(why error) {
			_, _ = fmt.Fprintln(said, "miso: without vsock the builder VM is reached over a virtio port:", why)
		},
	}

	var socket func() (*os.File, error)

	namespace, opening := vsockns.Open()
	switch {
	case errors.Is(opening, vsockns.ErrNotPrivate):
		// the driver then reaches the VM over its virtio port and says why
		driver.OpenVsock = func() (*os.File, error) { return nil, opening }
	case opening != nil:
		return opening
	default:
		defer func() { _ = namespace.Close() }()

		driver.OpenVsock = namespace.Device
		socket = namespace.Socket
	}

	vm, err := driver.Start(running, machine)
	if err != nil {
		return err
	}

	dial, err := reach.Agent(vm, socket)
	if err == nil {
		err = work(vm, dial)
	}

	stop()
	<-vm.Done()

	if err != nil {
		return fmt.Errorf("%w (the builder VM's console is in %s)", err, log)
	}

	return nil
}
