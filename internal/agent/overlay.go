package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"syscall"

	"golang.org/x/sys/unix"
)

// overlaid mounts layers, lowest first, on top of a bottom made in scratch
// and under a directory, and hands the root to work, whose writes go into
// the directory. The directory is no longer mounted once it returns. Work
// tells whether what it wrote is kept.
func (a *Agent) overlaid(layers []string, bottom func(scratch string) (string, error), upper string, work func(root string) (keep bool, err error)) error {
	// overlay wants its work directory on the file system of the upper one
	scratch, err := a.layers.Scratch()
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	root := filepath.Join(scratch, "root")
	overlayWork := filepath.Join(scratch, "work")
	for _, dir := range []string{root, overlayWork} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			return err
		}
	}

	// overlay takes the top layer first
	lowers := make([]string, 0, len(layers))
	for _, key := range slices.Backward(layers) {
		lowers = append(lowers, a.layers.Path(key))
	}

	// overlay needs a layer below even on scratch
	lowest, err := bottom(scratch)
	if err != nil {
		return err
	}

	lowers = append(lowers, lowest)

	if err := mountOverlay(root, lowers, upper, overlayWork); err != nil {
		return err
	}

	// removing scratch must never reach into the root, so it goes first
	defer func() { _ = syscall.Unmount(root, syscall.MNT_DETACH) }()

	keep, err := work(root)
	if err != nil || !keep {
		return err
	}

	// not detached, a busy root means something still writes into the layer
	return syscall.Unmount(root, 0)
}

// mountOverlay mounts on root the layers below, top first, with what is
// written going into upper.
func mountOverlay(root string, lowers []string, upper, work string) error {
	// overlay shows the top of the upper directory as /
	below, err := os.Stat(lowers[0])
	if err != nil {
		return err
	}

	if err := os.Chmod(upper, below.Mode().Perm()); err != nil {
		return err
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
		[2]string{"workdir", work},
		[2]string{"redirect_dir", "off"},
		[2]string{"metacopy", "off"},
	)

	overlay, err := unix.Fsopen("overlay", unix.FSOPEN_CLOEXEC)
	if err != nil {
		return fmt.Errorf("open overlay: %w", err)
	}

	defer func() { _ = unix.Close(overlay) }()

	for _, option := range options {
		if err := unix.FsconfigSetString(overlay, option[0], option[1]); err != nil {
			return fmt.Errorf("overlay option %s=%s: %w", option[0], option[1], err)
		}
	}

	if err := unix.FsconfigCreate(overlay); err != nil {
		return fmt.Errorf("create overlay: %w", err)
	}

	mount, err := unix.Fsmount(overlay, unix.FSMOUNT_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("mount overlay: %w", err)
	}

	err = unix.MoveMount(mount, "", unix.AT_FDCWD, root, unix.MOVE_MOUNT_F_EMPTY_PATH)
	// an open mount keeps the root busy, and a strict unmount then fails
	_ = unix.Close(mount)
	if err != nil {
		return fmt.Errorf("mount overlay on %s: %w", root, err)
	}

	return nil
}
