package check

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/The127/miso/internal/firmware"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
)

// Boot is an image booted for its checks.
type Boot struct {
	Driver    qemu.Driver
	Namespace Namespace

	// the firmware that boots the image from its disk, which a Kernel
	// replaces
	Firmware firmware.Firmware
	Image    qemu.Disk

	// a kernel and initrd that boot the image as the root of its disk,
	// without a firmware, on the microvm board
	Kernel *qemu.Kernel

	// holds what the firmware and the boot write, a boot's own
	Dir     string
	Console io.Writer

	// how long the boot may take, which callers must state. A firmware
	// that finds nothing to boot waits in its menu and never stops the VM
	Patience time.Duration
}

// Run boots the image and runs each check in it, one after the other.
func (b Boot) Run(ctx context.Context, checks []string) ([]Result, error) {
	if b.Namespace == nil {
		return nil, ErrNoNamespace
	}

	method, err := b.method()
	if err != nil {
		return nil, err
	}

	notices, err := b.listen()
	if err != nil {
		return nil, err
	}

	// ends the taking of notices after the boot
	defer func() { _ = notices.Close() }()

	ctx, stop := context.WithCancel(ctx)
	defer stop()

	vm, onPort, err := b.start(ctx, method, notices.Port())
	if err != nil {
		return nil, err
	}

	// QEMU writes to the console and into Dir until it is gone, and the
	// caller may read the one and remove the other once Run returns
	defer func() {
		stop()
		<-vm.Done()
	}()

	// the checks talk over vsock only, a VM on a port would never be heard
	if onPort != nil {
		return nil, fmt.Errorf("the image cannot be checked without vsock: %w", onPort)
	}

	if err := b.awaitBoot(ctx, vm, notices); err != nil {
		return nil, err
	}

	return b.runChecks(vm, checks)
}

// start is the VM the image boots in, told where to send its notices.
func (b Boot) start(ctx context.Context, method qemu.Boot, notifyPort uint32) (vm *qemu.VM, onPort, err error) {
	driver := b.Driver
	driver.OpenVsock = b.Namespace.Device
	// a VM on a port fails the boot, it is not only a notice to pass on
	driver.WithoutVsock = func(why error) { onPort = why }

	vm, err = driver.Start(ctx, b.machine(method, notifyPort))

	return vm, onPort, err
}

// machine is the VM the image boots in. An image booted from a kernel and
// its initrd is booted as Firecracker boots a microVM, on the microvm board.
func (b Boot) machine(method qemu.Boot, notifyPort uint32) qemu.Machine {
	return qemu.Machine{
		Boot:        method,
		Microvm:     b.Kernel != nil,
		MemoryMiB:   2048,
		CPUs:        2,
		Disks:       []qemu.Disk{b.Image},
		Console:     b.Console,
		Credentials: credentials(notifyPort),
		Temp:        b.Dir,
	}
}

// awaitBoot waits until the image has booted, the VM stopped, the caller
// gave up or the patience ran out.
func (b Boot) awaitBoot(ctx context.Context, vm *qemu.VM, notices *vsock.Listener) error {
	stopped, stop := context.WithCancelCause(ctx)
	defer stop(nil)

	wait, cancel := context.WithTimeoutCause(stopped, b.Patience, fmt.Errorf("the image did not boot within %s", b.Patience))
	defer cancel()

	// a VM that stops ends the wait for its boot, the notices can come from
	// nobody else
	go func() {
		select {
		case <-vm.Done():
			stop(fmt.Errorf("the image did not boot: %w", vm.Err()))
		case <-wait.Done():
		}
	}()

	closing := context.AfterFunc(wait, func() { _ = notices.Close() })

	err := booted(wait, notices, vm.CID())

	// the notices after the boot must stay open, and a wait that ended as
	// the boot came has closed them already
	if err == nil && closing() {
		return nil
	}

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if wait.Err() != nil {
		return context.Cause(wait)
	}

	return fmt.Errorf("the image did not boot: %w", err)
}

// runChecks runs each check on a connection of its own.
func (b Boot) runChecks(vm *qemu.VM, checks []string) ([]Result, error) {
	var results []Result
	for _, check := range checks {
		conn, err := b.dial(vm.CID(), shellPort)
		if err != nil {
			return nil, err
		}

		result, err := run(conn, check)
		_ = conn.Close()

		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// method is how the machine starts: the kernel when there is one, the
// firmware otherwise.
func (b Boot) method() (qemu.Boot, error) {
	if b.Kernel != nil {
		return *b.Kernel, nil
	}

	return b.flash()
}

// flash writes the firmware where QEMU reads it, the vars a copy of their
// own.
func (b Boot) flash() (qemu.Firmware, error) {
	flash := qemu.Firmware{Code: filepath.Join(b.Dir, "code.fd"), Vars: filepath.Join(b.Dir, "vars.fd")}
	if err := os.WriteFile(flash.Code, b.Firmware.Code, 0o600); err != nil {
		return qemu.Firmware{}, err
	}

	if err := os.WriteFile(flash.Vars, b.Firmware.Vars, 0o600); err != nil {
		return qemu.Firmware{}, err
	}

	return flash, nil
}
