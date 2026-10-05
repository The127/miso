package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/builder"
	"github.com/The127/miso/internal/check"
	"github.com/The127/miso/internal/download"
	"github.com/The127/miso/internal/firmware"
	"github.com/The127/miso/internal/qemu"
	"github.com/The127/miso/internal/vsockns"
)

// bootPatience is how long a checked disk may take to boot, far more than
// an image with a stock kernel needs.
const bootPatience = 5 * time.Minute

// checker boots each checked output on a vsock of its own and runs its checks:
// a rootfs with its kernel and initrd, any other output under OVMF. The boots
// write their scratch into dir and their console to console.
func checker(ctx context.Context, blobs *download.Store, dir string, console io.Writer) builder.Check {
	return func(disk string, request build.Request) error {
		namespace, err := vsockns.Open()
		if errors.Is(err, vsockns.ErrNotPrivate) {
			return fmt.Errorf("CHECK needs a vsock of its own: %w", err)
		}

		if err != nil {
			return err
		}

		defer func() { _ = namespace.Close() }()

		scratch, err := os.MkdirTemp(dir, "check-")
		if err != nil {
			return err
		}

		defer func() { _ = os.RemoveAll(scratch) }()

		boot := check.Boot{
			Namespace: namespace,
			// what the boot writes must not reach the disk that is delivered
			Image:    qemu.Disk{Path: disk, Format: "raw", Serial: "image", Access: qemu.Snapshot, CD: request.CD},
			Dir:      scratch,
			Console:  console,
			Patience: bootPatience,
		}

		if err := bootsFrom(ctx, blobs, disk, request, &boot); err != nil {
			return err
		}

		results, err := boot.Run(ctx, commandsOf(request))
		if err != nil {
			return err
		}

		for i, result := range results {
			if result.Code != 0 {
				return request.CheckFailed(i, result.Code, result.Output)
			}
		}

		return nil
	}
}

// bootsFrom sets what boots the checked output: the kernel and initrd beside
// a rootfs, the firmware for any other output.
func bootsFrom(ctx context.Context, blobs *download.Store, disk string, request build.Request, boot *check.Boot) error {
	if request.Boot == nil {
		found, err := firmware.Ready(ctx, blobs)
		if err != nil {
			return err
		}

		boot.Firmware = found

		return nil
	}

	outputs := filepath.Dir(disk)
	boot.Kernel = &qemu.Kernel{
		Image:       filepath.Join(outputs, request.Boot.Kernel),
		Initramfs:   filepath.Join(outputs, request.Boot.Initrd),
		CommandLine: request.Boot.Cmdline,
	}

	return nil
}

// commandsOf are the commands of the checks of a request, in their order.
func commandsOf(request build.Request) []string {
	commands := make([]string, 0, len(request.Checks))
	for _, c := range request.Checks {
		commands = append(commands, c.Command)
	}

	return commands
}
