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
func diskRequest(step plan.Step, instruction imagefile.Output, under rootfs, tools rootfs, partitions []protocol.Partition) (protocol.Disk, error) {
	var refused error
	unknown, hasUnknown := unknownOption(instruction.Options)
	switch {
	case instruction.Kind != "disk" && instruction.Kind != "iso":
		refused = ErrUnknownKind
	case hasUnknown:
		refused = fmt.Errorf("--%s: %w", unknown, ErrUnknownOption)
	}

	if refused != nil {
		line, written := imagefile.Written(instruction)

		return protocol.Disk{}, &imagefile.Error{Line: line, Err: fmt.Errorf("%s: %w", written, refused)}
	}

	return protocol.Disk{Key: step.Key, Layers: under.layers, Tools: tools.layers, ElTorito: instruction.Kind == "iso", Partitions: slices.Clone(partitions)}, nil
}

// unknownOption is the first option in sorted order, so the same build file
// always names the same one. A disk takes none.
func unknownOption(options map[string]string) (string, bool) {
	for _, name := range slices.Sorted(maps.Keys(options)) {
		return name, true
	}

	return "", false
}

func partitionOf(partition imagefile.Partition) protocol.Partition {
	settings := make([]protocol.Setting, 0, len(partition.Settings))
	for _, setting := range partition.Settings {
		settings = append(settings, protocol.Setting{Key: setting.Key, Value: setting.Value})
	}

	return protocol.Partition{Name: partition.Name, Settings: settings}
}
