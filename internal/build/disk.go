package build

import (
	"fmt"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// diskRequest is what an output asks for on the root file system under it,
// with the tools taken from where their stage ends.
func diskRequest(step plan.Step, instruction imagefile.Output, under rootfs, tools rootfs) (protocol.Disk, error) {
	var refused error
	switch _, hasTools := instruction.Options[plan.Tools]; {
	case instruction.Kind != "disk":
		refused = ErrUnknownKind
	case !hasTools:
		refused = ErrNoTools
	}

	if refused != nil {
		line, written := imagefile.Written(instruction)

		return protocol.Disk{}, &imagefile.Error{Line: line, Err: fmt.Errorf("%s: %w", written, refused)}
	}

	return protocol.Disk{Key: step.Key, Layers: under.layers, Tools: tools.layers}, nil
}
