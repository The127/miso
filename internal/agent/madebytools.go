package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/The127/miso/internal/sandbox"
)

// toolsDir is what the tools of a stage are shown to make a file from the
// image: the image mounted read-only, a directory for the booting mount,
// and the directory the file goes into.
type toolsDir struct {
	scratch string
	image   mountedImage
	booting string
	output  string
}

// toolsRun is what the tools do in their root, and the code they exit with.
type toolsRun func(root string) (int, error)

// madeByTools keeps what the tools of a stage write, in a root of their own,
// as the layer of a key. The noun is used in the error text. A key whose layer
// is there already has it.
func (a *Agent) madeByTools(key string, layers, tools []string, noun string, prepare func(toolsDir) (toolsRun, error)) error {
	there, at, err := a.found(key, slices.Concat(layers, tools))
	if err != nil || there {
		return err
	}

	scratch, err := a.layers.Scratch()
	if err != nil {
		return err
	}

	defer func() { _ = os.RemoveAll(scratch) }()

	image, err := a.mountImage(scratch, layers)
	if err != nil {
		return err
	}

	// removing scratch must never reach into the image, so it goes first
	defer image.unmount()

	work, err := a.layers.Begin(key)
	if err != nil {
		return err
	}

	code, err := a.inTools(scratch, image, work.Dir(), tools, prepare)
	if err == nil && code != 0 {
		err = fmt.Errorf("the tools failed making the %s: exit code %d", noun, code)
	}

	if err != nil {
		_ = work.Discard()

		return err
	}

	return work.FinishAt(at)
}

// inTools has the tools run what prepare makes ready, as a RUN does, but
// nothing they write is kept.
func (a *Agent) inTools(scratch string, image mountedImage, output string, tools []string, prepare func(toolsDir) (toolsRun, error)) (int, error) {
	booting := filepath.Join(scratch, "booting")
	if err := os.Mkdir(booting, 0o700); err != nil {
		return 0, err
	}

	upper := filepath.Join(scratch, "tools")
	if err := os.Mkdir(upper, 0o700); err != nil {
		return 0, err
	}

	run, err := prepare(toolsDir{scratch: scratch, image: image, booting: booting, output: output})
	if err != nil {
		return 0, err
	}

	code := 0
	err = a.overlaid(tools, sandbox.Floor, nil, upper, func(root string) (bool, error) {
		var err error
		code, err = run(root)

		return false, err
	})

	return code, err
}
