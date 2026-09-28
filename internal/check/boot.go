package check

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

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

	// holds what the firmware writes, a boot's own
	Dir     string
	Console io.Writer
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

	vm, err := b.Driver.Start(ctx, qemu.Machine{
		Firmware:    flash,
		MemoryMiB:   2048,
		CPUs:        2,
		Disks:       []qemu.Disk{b.Image},
		Console:     b.Console,
		Credentials: Credentials(notices.Port()),
	})
	if err != nil {
		return nil, err
	}

	// a VM that stops ends the wait for its boot, the notices can come from
	// nobody else
	go func() {
		select {
		case <-vm.Done():
		case <-ctx.Done():
		}

		_ = notices.Close()
	}()

	if err := Booted(notices); err != nil {
		return nil, notBooted(vm, err)
	}

	var results []Result
	for _, check := range checks {
		conn, err := vsock.Dial(vm.CID(), shellPort)
		if err != nil {
			return nil, err
		}

		result, err := Run(conn, check)
		_ = conn.Close()

		if err != nil {
			return nil, err
		}

		results = append(results, result)
	}

	return results, nil
}

// notBooted is why a wait for the boot ended without it, the VM's own
// reason once it stopped.
func notBooted(vm *qemu.VM, err error) error {
	select {
	case <-vm.Done():
		return fmt.Errorf("the image did not boot: %w", vm.Err())
	default:
		return fmt.Errorf("the image did not boot: %w", err)
	}
}

// flash writes the firmware where QEMU reads it, the vars a copy of their
// own.
func (b Boot) flash() (*qemu.Firmware, error) {
	flash := &qemu.Firmware{Code: filepath.Join(b.Dir, "code.fd"), Vars: filepath.Join(b.Dir, "vars.fd")}
	if err := os.WriteFile(flash.Code, b.Firmware.Code, 0o600); err != nil {
		return nil, err
	}

	if err := os.WriteFile(flash.Vars, b.Firmware.Vars, 0o600); err != nil {
		return nil, err
	}

	return flash, nil
}
