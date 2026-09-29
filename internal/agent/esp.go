package agent

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
)

// makeESP makes in a directory the layer that lies on the image, mounted at
// root, with the ESP as its efi, where the image's repart definitions take
// it from, and with the image's systemd-boot as the ESP's fallback loader.
func makeESP(image fs.FS, root, esp string) error {
	fallback := filepath.Join(esp, "efi", "EFI", "BOOT")
	if err := os.MkdirAll(fallback, 0o755); err != nil { //nolint:gosec // an image's directories are open to all
		return err
	}

	if err := os.Mkdir(filepath.Join(esp, "efi", "EFI", "Linux"), 0o755); err != nil { //nolint:gosec // an image's directories are open to all
		return err
	}

	// the image's root shows the mode and owner of the top layer, which the
	// ESP is
	info, err := os.Stat(root)
	if err != nil {
		return err
	}

	if err := os.Chmod(esp, info.Mode().Perm()); err != nil {
		return err
	}

	owner, _ := info.Sys().(*syscall.Stat_t)
	if err := os.Lchown(esp, int(owner.Uid), int(owner.Gid)); err != nil {
		return err
	}

	err = copyPart(image, systemdBoot, filepath.Join(fallback, "BOOTX64.EFI"))
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("the image has no systemd-boot at /%s", systemdBoot)
	}

	return err
}
