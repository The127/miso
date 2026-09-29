package build

import (
	"fmt"
	"maps"
	"slices"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// diskRequest is what an output of a disk or an ISO asks for on the root
// file system under it, with the tools taken from where their stage ends.
func diskRequest(step plan.Step, instruction imagefile.Output, under rootfs, tools rootfs) (protocol.Disk, error) {
	var refused error
	_, hasTools := instruction.Options[plan.Tools]
	unknown, hasUnknown := unknownOption(instruction.Options)
	switch {
	case instruction.Kind != "disk" && instruction.Kind != "iso":
		refused = ErrUnknownKind
	case hasUnknown:
		refused = fmt.Errorf("--%s: %w", unknown, ErrUnknownOption)
	case !hasTools:
		refused = ErrNoTools
	}

	if refused != nil {
		line, written := imagefile.Written(instruction)

		return protocol.Disk{}, &imagefile.Error{Line: line, Err: fmt.Errorf("%s: %w", written, refused)}
	}

	return protocol.Disk{Key: step.Key, Layers: under.layers, Tools: tools.layers, ElTorito: instruction.Kind == "iso"}, nil
}

// unknownOption is the first option in sorted order that a disk does not
// take, so the same build file always names the same one.
func unknownOption(options map[string]string) (string, bool) {
	for _, name := range slices.Sorted(maps.Keys(options)) {
		if name != plan.Tools {
			return name, true
		}
	}

	return "", false
}
