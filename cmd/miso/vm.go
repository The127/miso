package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/reach"
)

// builderArch is the architecture of the builder VM, always the host's, so
// it runs on KVM. Until cross builds, images are built only for it.
const builderArch = runtime.GOARCH

// hostArches are the hosts miso has a builder kernel, a firmware and base
// images for.
var hostArches = []string{"amd64", "arm64"}

// hostRefusal is why miso cannot run a builder on host, or nil when it can.
func hostRefusal(host string) error {
	if !slices.Contains(hostArches, host) {
		return fmt.Errorf("miso builds on amd64 and arm64, not on %s", host)
	}

	return nil
}

// bootFiles writes the builder kernel and an initramfs with this very miso
// as its init into a directory of their own.
func bootFiles(ctx context.Context, blobs *download.Store) (builder.Boot, error) {
	kernel, err := builderkernel.Ready(ctx, blobs, builderArch)
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

	boot, err := builder.WriteBoot(dir, builderArch, kernel, init)
	if err != nil {
		return builder.Boot{}, errors.Join(err, os.RemoveAll(dir))
	}

	return boot, nil
}

// inBuilder does the work with the builder VM booted, its console in a log
// file, and stops the VM afterwards. The work dials the VM's agent. What the
// user should know about how it runs goes to said.
func inBuilder(ctx context.Context, machine qemu.Machine, log string, said io.Writer, work inVM) error {
	console, err := os.Create(log)
	if err != nil {
		return err
	}

	defer func() { _ = console.Close() }()

	machine.Console = console

	running, stop := context.WithCancel(ctx)
	defer stop()

	driver := qemu.Driver{
		WithoutKVM: func(why error) {
			_, _ = fmt.Fprintln(said, "miso: without KVM the builder VM runs much slower:", why)
		},
		WithoutVsock: func(why error) {
			_, _ = fmt.Fprintln(said, "miso: without vsock the builder VM is reached over a virtio port:", why)
		},
	}

	socket, closeVsock, err := privateVsock(&driver)
	if err != nil {
		return err
	}

	defer closeVsock()

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
