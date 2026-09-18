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

	code, err := a.runOn(work.Dir(), run)
	if err != nil || code != 0 {
		_ = work.Discard()

		return code, err
	}

	return 0, work.Finish()
}

// runOn runs a command on top of layers with what it writes going into a
// directory, which is no longer mounted once it returns.
func (a *Agent) runOn(upper string, run protocol.Run) (int, error) {
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

	// overlay takes the top layer first
	lowers := make([]string, 0, len(run.Layers))
	for _, key := range slices.Backward(run.Layers) {
		lowers = append(lowers, a.layers.Path(key))
	}

	options := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", strings.Join(lowers, ":"), upper, overlayWork)
	if err := syscall.Mount("overlay", root, "overlay", 0, options); err != nil {
		return 0, fmt.Errorf("mount overlay on %s: %w", root, err)
	}

	// removing scratch must never reach into the root, so it goes first
	defer func() { _ = syscall.Unmount(root, syscall.MNT_DETACH) }()

	cmd := exec.Command("/bin/sh", "-c", run.Command) //nolint:gosec // running what the build file says is what a RUN is
	cmd.SysProcAttr = &syscall.SysProcAttr{Chroot: root}
	cmd.Dir = "/"
	err = cmd.Run()
	if exited, ok := errors.AsType[*exec.ExitError](err); ok {
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
