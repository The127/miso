package agent

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// mountOverlay mounts on root the layers below, top first, with what is
// written going into upper.
func mountOverlay(root string, lowers []string, upper, work string) error {
	// overlay shows the top of the upper directory as /
	if len(lowers) > 0 {
		below, err := os.Stat(lowers[0])
		if err != nil {
			return err
		}

		if err := os.Chmod(upper, below.Mode().Perm()); err != nil {
			return err
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
