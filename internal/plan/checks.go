package plan

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/The127/miso/internal/imagefile"
)

// ErrNothingToCheck is a CHECK with no OUTPUT before it in its stage.
var ErrNothingToCheck = errors.New("no OUTPUT before it")

// ErrNotBooted is a CHECK after an output that is never booted.
var ErrNotBooted = errors.New("is never booted")

// ErrNoBootFiles is a CHECK of a rootfs that has no kernel or initrd above
// it to boot with.
var ErrNoBootFiles = errors.New("has no kernel or initrd to boot with")

// ErrNoRoot is a CHECK of a rootfs whose kernel command line says nothing of
// its root. A kernel with no root= hangs until the boot gives up.
var ErrNoRoot = errors.New("is booted with no root= on its kernel command line")

// bootsWith are the outputs a rootfs must have above it to be booted.
var bootsWith = []string{KindKernel, KindInitrd}

func checkChecks(stage imagefile.Stage) error {
	var last *imagefile.Output

	// the kinds of output above the last one
	var above, seen, cmdline, cmdlineAbove []string
	for _, instruction := range stage.Instructions {
		switch step := instruction.(type) {
		case imagefile.Output:
			last = &step
			above = slices.Clone(seen)
			cmdlineAbove = slices.Clone(cmdline)
			seen = append(seen, step.Kind)
		case imagefile.Cmdline:
			cmdline = append(cmdline, step.Text)
		case imagefile.Check:
			if last == nil {
				return at(step.Line, fmt.Errorf("CHECK: %w", ErrNothingToCheck))
			}

			if NeverBooted(last.Kind) {
				return at(step.Line, fmt.Errorf("CHECK: the %s %s %w", last.Kind, last.Name, ErrNotBooted))
			}

			if last.Kind == KindRootfs {
				for _, kind := range bootsWith {
					if !slices.Contains(above, kind) {
						return at(step.Line, fmt.Errorf("CHECK: the rootfs %s: no OUTPUT %s above it, it %w", last.Name, kind, ErrNoBootFiles))
					}
				}

				if !namesRoot(cmdlineAbove) {
					return at(step.Line, fmt.Errorf("CHECK: the rootfs %s %w, add a CMDLINE with root= above it", last.Name, ErrNoRoot))
				}
			}
		}
	}

	return nil
}

// namesRoot says whether kernel command line words name the root.
func namesRoot(lines []string) bool {
	for _, line := range lines {
		for _, word := range strings.Fields(line) {
			if strings.HasPrefix(word, "root=") {
				return true
			}
		}
	}

	return false
}
