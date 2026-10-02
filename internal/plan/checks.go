package plan

import (
	"errors"
	"fmt"
	"slices"

	"github.com/The127/miso/internal/imagefile"
)

// ErrNothingToCheck is a CHECK with no OUTPUT before it in its stage.
var ErrNothingToCheck = errors.New("no OUTPUT before it")

// ErrNotBooted is a CHECK after an output that is never booted.
var ErrNotBooted = errors.New("is never booted")

// ErrNoBootFiles is a CHECK of a rootfs that has no kernel or initrd above
// it to boot with.
var ErrNoBootFiles = errors.New("has no kernel or initrd to boot with")

// bootsWith are the outputs a rootfs must have above it to be booted.
var bootsWith = []string{KindKernel, KindInitrd}

func checkChecks(stage imagefile.Stage) error {
	var last *imagefile.Output

	// the kinds of output above the last one
	var above, seen []string
	for _, instruction := range stage.Instructions {
		switch step := instruction.(type) {
		case imagefile.Output:
			last = &step
			above = slices.Clone(seen)
			seen = append(seen, step.Kind)
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
			}
		}
	}

	return nil
}
