package contextfiles

import (
	"errors"
	"fmt"
	"io"

	"github.com/The127/miso/internal/buildcontext"
	"github.com/The127/miso/internal/protocol"
)

// ErrFromAStage is a copy from an earlier stage, whose files are in that
// stage's layers and never in the build context.
var ErrFromAStage = errors.New("a copy from a stage takes nothing from the build context")

// ErrNotPlanned is a copy that does not say what the plan said of each of
// its sources, so nothing it reads could be checked against the plan.
var ErrNotPlanned = errors.New("the plan says nothing of some sources")

// Of are the files of a copy, read from the build context.
func Of(context *buildcontext.Dir) func(protocol.Copy) protocol.Files {
	return func(request protocol.Copy) protocol.Files {
		return func(send func(protocol.Entry, io.Reader) error) error {
			if request.Stage != "" {
				return fmt.Errorf("COPY --from=%s: %w", request.Stage, ErrFromAStage)
			}

			if len(request.Digests) < len(request.Sources) {
				return fmt.Errorf("sources: %d, digests: %d: %w", len(request.Sources), len(request.Digests), ErrNotPlanned)
			}

			for source, name := range request.Sources {
				err := context.Pack(name, request.Digests[source], func(entry buildcontext.Entry, content io.Reader) error {
					return send(protocol.Entry{Source: source, Kind: entry.Kind, Path: entry.Path, Mode: entry.Mode, Target: entry.Target, Size: entry.Size}, content)
				})
				if err != nil {
					return err
				}
			}

			return nil
		}
	}
}
