package agent

import (
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/The127/miso/internal/fstab"
)

// mountSubmounts mounts the submounts of a root file system on a device
// read-only into the root mounted on a directory.
func mountSubmounts(device, kind, target string, submounts []fstab.Entry) error {
	image, err := os.OpenRoot(target)
	if err != nil {
		return err
	}

	defer func() { _ = image.Close() }()

	for _, submount := range submounts {
		// the other options tune a running system and mean nothing to a copy
		var data string

		for _, option := range submount.Options {
			if strings.HasPrefix(option, "subvol=") {
				data = option
			}
		}

		// opened within the image, so a link cannot lead the mount out of it
		point, err := image.Open(strings.TrimPrefix(submount.Target, "/"))
		if err != nil {
			return err
		}

		err = syscall.Mount(device, fmt.Sprintf("/proc/self/fd/%d", point.Fd()), kind, syscall.MS_RDONLY, data)
		_ = point.Close()

		if err != nil {
			return err
		}
	}

	return nil
}
