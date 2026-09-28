package check

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/firmware"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsock"
)

// Boot is an image booted for its checks.
type Boot struct {
	Driver   qemu.Driver
	Firmware firmware.Firmware
	Image    qemu.Disk

	// holds what the firmware and the boot write, a boot's own
	Dir     string
	Console io.Writer

	// how long the boot may take, which callers must state. A firmware
	// that finds nothing to boot waits in its menu and never stops the VM
	Patience time.Duration
}

// Run boots the image and runs each check in it, one after the other.
func (b Boot) Run(ctx context.Context, checks []string) ([]Result, error) {
	flash, err := b.flash()
	if err != nil {
		return nil, err
	}

	notices, err := vsock.Listen(unix.VMADDR_PORT_ANY)
	if err != nil {
		return nil, err
	}

	// ends the taking of notices after the boot
	defer func() { _ = notices.Close() }()

	ctx, stop := context.WithCancel(ctx)
	defer stop()

	vm, err := b.start(ctx, flash, notices.Port())
	if err != nil {
		return nil, err
	}

	// QEMU writes to the console and into Dir until it is gone, and the
	// caller may read the one and remove the other once Run returns
	defer func() {
		stop()
		<-vm.Done()
	}()

	if err := b.awaitBoot(ctx, vm, notices); err != nil {
		return nil, err
	}

	return runChecks(vm, checks)
}

// start is the VM the image boots in, told where to send its notices.
func (b Boot) start(ctx context.Context, flash qemu.Firmware, notifyPort uint32) (*qemu.VM, error) {
	return b.Driver.Start(ctx, qemu.Machine{
		Boot:        flash,
		MemoryMiB:   2048,
		CPUs:        2,
		Disks:       []qemu.Disk{b.Image},
		Console:     b.Console,
		Credentials: credentials(notifyPort),
		Temp:        b.Dir,
	})
}

// awaitBoot waits until the image has booted, the VM stopped, the caller
// gave up or the patience ran out.
func (b Boot) awaitBoot(ctx context.Context, vm *qemu.VM, notices *vsock.Listener) error {
	// a VM that stops ends the wait for its boot, the notices can come from
	// nobody else
	go func() {
		select {
		case <-vm.Done():
		case <-ctx.Done():
		}

		_ = notices.Close()
	}()

	late := make(chan struct{})
	patience := time.AfterFunc(b.Patience, func() {
		close(late)

		_ = notices.Close()
	})
	defer patience.Stop()

	if err := booted(notices); err != nil {
		return b.notBooted(ctx, vm, late, err)
	}

	// the notices after the boot were closed with it. Stop only says the
	// timer started, so it is waited for
	if !patience.Stop() {
		<-late

		return b.notBooted(ctx, vm, late, nil)
	}

	return nil
}

// runChecks runs each check on a connection of its own.
func runChecks(vm *qemu.VM, checks []string) ([]Result, error) {
	var results []Result
	for _, check := range checks {
		conn, err := vsock.Dial(vm.CID(), shellPort)
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

// notBooted is why a wait for the boot ended without it: the caller gave
// up, the patience ran out, or the VM's own reason once it stopped.
func (b Boot) notBooted(ctx context.Context, vm *qemu.VM, late <-chan struct{}, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	select {
	case <-late:
		return fmt.Errorf("the image did not boot within %s", b.Patience)
	default:
	}

	select {
	case <-vm.Done():
		return fmt.Errorf("the image did not boot: %w", vm.Err())
	default:
		return fmt.Errorf("the image did not boot: %w", err)
	}
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
