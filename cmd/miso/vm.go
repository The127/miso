package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/qemu"
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
// file, and stops the VM afterwards.
func inBuilder(ctx context.Context, machine qemu.Machine, log string, work func(vm *qemu.VM) error) error {
	console, err := os.Create(log)
	if err != nil {
		return err
	}

	defer func() { _ = console.Close() }()

	machine.Console = console

	running, stop := context.WithCancel(ctx)
	defer stop()

	vm, err := qemu.Driver{Binary: "qemu-system-x86_64"}.Start(running, machine)
	if err != nil {
		return err
	}

	err = work(vm)
	stop()
	<-vm.Done()

	if err != nil {
		return fmt.Errorf("%w (the builder VM's console is in %s)", err, log)
	}

	return nil
}
