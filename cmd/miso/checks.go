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

// checker boots each checked disk under OVMF on a vsock of its own and runs
// its checks. The boots write their scratch into dir and their console to
// console.
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

		commands := make([]string, 0, len(request.Checks))
		for _, c := range request.Checks {
			commands = append(commands, c.Command)
		}

		boot := check.Boot{
			Driver:    qemu.Driver{Binary: "qemu-system-x86_64"},
			Namespace: namespace,
			// what the boot writes must not reach the disk that is delivered
			Image:    qemu.Disk{Path: disk, Format: "raw", Serial: "image", Access: qemu.Snapshot, CD: request.CD},
			Dir:      scratch,
			Console:  console,
			Patience: bootPatience,
		}

		if request.Boot != nil {
			// the outputs of the kernel and the initrd are beside the rootfs
			outputs := filepath.Dir(disk)
			boot.Kernel = &qemu.Kernel{
				Image:       filepath.Join(outputs, request.Boot.Kernel),
				Initramfs:   filepath.Join(outputs, request.Boot.Initrd),
				CommandLine: request.Boot.Cmdline,
			}
		} else {
			boot.Firmware, err = firmware.Ready(ctx, blobs)
			if err != nil {
				return err
			}
		}

		results, err := boot.Run(ctx, commands)
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
