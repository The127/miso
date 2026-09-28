package build

import (
	"fmt"
	"strings"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// copyRequest is what a copy asks for on the root file system under it,
// with the sources of a stage taken from where that stage ends. A copy of
// an output is refused.
func copyRequest(step plan.Step, instruction imagefile.Copy, under rootfs, stage rootfs) (protocol.Copy, error) {
	copying := protocol.Copy{Key: step.Key, Layers: under.layers, Sources: instruction.Sources, Destination: instruction.Destination}
	if instruction.From == "" {
		for _, file := range step.Files {
			copying.Digests = append(copying.Digests, file.Digest)
		}

		return copying, nil
	}

	for _, source := range instruction.Sources {
		if !strings.HasPrefix(source, "/") {
			line, written := imagefile.Written(instruction)

			return protocol.Copy{}, &imagefile.Error{Line: line, Err: fmt.Errorf("%s: %w", written, ErrOutputNotBuilt)}
		}
	}

	copying.Stage = instruction.From
	copying.From = stage.layers

	return copying, nil
}
