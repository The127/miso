package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"

	"golang.org/x/sys/unix"

	"github.com/The127/miso/internal/protocol"
)

// Run runs a command on top of layers and keeps what it writes as the layer
// of the key.
func (a *Agent) Run(ctx context.Context, run protocol.Run, out io.Writer) (int, error) {
	work, err := a.layers.Begin(run.Key)
	if err != nil {
		return 0, err
	}

	code, err := a.runOn(ctx, work.Dir(), run, out)
	if err != nil || code != 0 {
		_ = work.Discard()

		return code, err
	}

	return 0, work.Finish()
}

// runOn runs a command on top of layers with what it writes going into a
// directory, which is no longer mounted once it returns.
func (a *Agent) runOn(ctx context.Context, upper string, run protocol.Run, out io.Writer) (int, error) {
	// overlay wants its work directory on the file system of the upper one
	scratch, err := a.layers.Scratch()
	if err != nil {
		return 0, err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	root := filepath.Join(scratch, "root")
	overlayWork := filepath.Join(scratch, "work")
	for _, dir := range []string{root, overlayWork} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			return 0, err
		}
	}

	// overlay takes the top layer first
	lowers := make([]string, 0, len(run.Layers))
	for _, key := range slices.Backward(run.Layers) {
		lowers = append(lowers, a.layers.Path(key))
	}

	// overlay shows the top of the upper directory as /
	if len(lowers) > 0 {
		below, err := os.Stat(lowers[0])
		if err != nil {
			return 0, err
		}

		if err := os.Chmod(upper, below.Mode().Perm()); err != nil {
			return 0, err
		}
	}

	// one option per layer, because all layers in one text outgrow the page
	// mount copies its options into
	options := make([][2]string, 0, len(lowers)+4)
	for _, lower := range lowers {
		options = append(options, [2]string{"lowerdir+", lower})
	}

	// a layer holds real files whatever the kernel's defaults, so it stays
	// whole once it leaves overlay
	options = append(options,
		[2]string{"upperdir", upper},
		[2]string{"workdir", overlayWork},
		[2]string{"redirect_dir", "off"},
		[2]string{"metacopy", "off"},
	)

	overlay, err := unix.Fsopen("overlay", unix.FSOPEN_CLOEXEC)
	if err != nil {
		return 0, fmt.Errorf("open overlay: %w", err)
	}

	defer func() { _ = unix.Close(overlay) }()

	for _, option := range options {
		if err := unix.FsconfigSetString(overlay, option[0], option[1]); err != nil {
			return 0, fmt.Errorf("overlay option %s=%s: %w", option[0], option[1], err)
		}
	}

	if err := unix.FsconfigCreate(overlay); err != nil {
		return 0, fmt.Errorf("create overlay: %w", err)
	}

	mount, err := unix.Fsmount(overlay, unix.FSMOUNT_CLOEXEC, 0)
	if err != nil {
		return 0, fmt.Errorf("mount overlay: %w", err)
	}

	err = unix.MoveMount(mount, "", unix.AT_FDCWD, root, unix.MOVE_MOUNT_F_EMPTY_PATH)
	// an open mount keeps the root busy for the strict unmount at the end
	_ = unix.Close(mount)
	if err != nil {
		return 0, fmt.Errorf("mount overlay on %s: %w", root, err)
	}

	// removing scratch must never reach into the root, so it goes first
	defer func() { _ = syscall.Unmount(root, syscall.MNT_DETACH) }()

	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", run.Command) //nolint:gosec // running what the build file says is what a RUN is
	// the shell is the init of its own PID namespace, and the kernel kills
	// whatever it leaves behind before the wait for it returns
	cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: root, Cloneflags: syscall.CLONE_NEWPID}
	cmd.Dir = "/"
	cmd.Stdout = out
	cmd.Stderr = out
	// Docker's defaults, and os/exec keeps the last of a key, so the build
	// file's own values win
	cmd.Env = append([]string{
		"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
		"HOME=/root",
	}, run.Env...)
	err = cmd.Run()
	// the kill that stops a cancelled command looks like an exit of its own
	if ctx.Err() != nil {
		return 0, ctx.Err()
	}

	if exited, ok := errors.AsType[*exec.ExitError](err); ok {
		// as a shell reports it, 137 for SIGKILL
		if status, ok := exited.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return 128 + int(status.Signal()), nil
		}

		return exited.ExitCode(), nil
	}

	if err != nil {
		return 0, err
	}

	// not detached, a busy root means something still writes into the layer
	if err := syscall.Unmount(root, 0); err != nil {
		return 0, err
	}

	return 0, nil
}
