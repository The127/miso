package agent

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// ukiFile is the name the UKI of a split disk is kept under, next to its
// partition files.
const ukiFile = "uki.efi"

// keepUKI copies the UKI the tools built into the ESP to the output. The
// tools wrote it, so it may be a link to anything in the builder, or a FIFO
// no one ever writes to.
func keepUKI(boot boot, output string) error {
	esp, err := os.OpenRoot(boot.esp)
	if err != nil {
		return err
	}

	defer func() { _ = esp.Close() }()

	uki, err := esp.OpenFile(filepath.Join("efi", "EFI", "Linux", boot.kernel.Version+".efi"), os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return err
	}

	defer func() { _ = uki.Close() }()

	info, err := uki.Stat()
	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {
		return fmt.Errorf("the UKI of %s is no regular file", boot.kernel.Version)
	}

	kept, err := os.OpenFile(filepath.Join(output, ukiFile), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644) //nolint:gosec // the output is a place of its own
	if err != nil {
		return err
	}

	if _, err := io.Copy(kept, uki); err != nil {
		return errors.Join(err, kept.Close())
	}

	return kept.Close()
}
