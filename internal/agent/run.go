package agent

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/The127/miso/internal/protocol"
)

// Run runs a command on top of layers and keeps what it writes as the layer
// of the key.
func (a *Agent) Run(_ context.Context, run protocol.Run, _ io.Writer) (int, error) {
	work, err := a.layers.Begin(run.Key)
	if err != nil {
		return 0, err
	}

	scratch, err := os.MkdirTemp(a.scratch, "run-")
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

	lowers := make([]string, 0, len(run.Layers))
	for _, key := range run.Layers {
		lowers = append(lowers, a.layers.Path(key))
	}

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(lowers, ":"), work.Dir(), overlayWork)
	if err := syscall.Mount("overlay", root, "overlay", 0, options); err != nil {
		return 0, fmt.Errorf("mount overlay on %s: %w", root, err)
	}

	cmd := exec.Command("/bin/sh", "-c", run.Command) //nolint:gosec // running what the build file says is what a RUN is
	cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: root}
	cmd.Dir = "/"
	if err := cmd.Run(); err != nil {
		return 0, err
	}

	if err := syscall.Unmount(root, 0); err != nil {
		return 0, err
	}

	return 0, work.Finish()
}
