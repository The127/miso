package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/builderkernel"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/emulator"
	"github.com/The127/miso/internal/initramfs"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/reach"
)

// builderArch is the architecture of the builder VM, always the host's, so
// it runs on KVM.
const builderArch = runtime.GOARCH

// arches are what miso has a builder kernel, a firmware, an emulator and
// base images for. It builds on each, for each.
var arches = []string{"amd64", "arm64"}

// hostRefusal is why miso cannot run a builder on host, or nil when it can.
func hostRefusal(host string) error {
	if !slices.Contains(arches, host) {
		return fmt.Errorf("miso builds on %s, not on %s", strings.Join(arches, " and "), host)
	}

	return nil
}

// emulatorsFor are what the builder VM needs to run the programs of a build
// for target, nothing for a build of its own architecture.
func emulatorsFor(target string) ([]initramfs.Emulator, error) {
	if target == builderArch {
		return nil, nil
	}

	from, program, err := emulator.Host(target)
	if err != nil {
		return nil, err
	}

	return []initramfs.Emulator{{Arch: target, From: from, Program: program}}, nil
}

// bootFiles writes the builder kernel and an initramfs with this very miso
// as its init, and what runs the programs of a build for target, into a
// directory of their own.
func bootFiles(ctx context.Context, blobs *download.Store, target string) (builder.Boot, error) {
	kernel, err := builderkernel.Ready(ctx, blobs, builderArch)
	if err != nil {
		return builder.Boot{}, err
	}

	emulators, err := emulatorsFor(target)
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

	boot, err := builder.WriteBoot(dir, builderArch, kernel, init, emulators...)
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
