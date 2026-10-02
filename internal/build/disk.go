package build

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
	"github.com/The127/miso/internal/protocol"
)

// diskRequest is what an output of a disk, an ISO or a rootfs asks for on the root
// file system under it, with the tools taken from where their stage ends.
func diskRequest(step plan.Step, instruction imagefile.Output, under rootfs, tools rootfs, inputs diskInputs) (protocol.Message, error) {
	var refused error
	unknown, hasUnknown := unknownOption(instruction.Options)
	switch {
	case instruction.Kind != "disk" && instruction.Kind != "iso" && instruction.Kind != "rootfs":
		refused = ErrUnknownKind
	case hasUnknown:
		refused = fmt.Errorf("--%s: %w", unknown, ErrUnknownOption)
	}

	if refused != nil {
		line, written := imagefile.Written(instruction)

		return nil, &imagefile.Error{Line: line, Err: fmt.Errorf("%s: %w", written, refused)}
	}

	if instruction.Kind == "rootfs" {
		return protocol.Rootfs{Key: step.Key, Layers: under.layers, Tools: tools.layers}, nil
	}

	return protocol.Disk{Key: step.Key, Layers: under.layers, Tools: tools.layers, ElTorito: instruction.Kind == "iso", Partitions: slices.Clone(inputs.partitions), Cmdline: strings.Join(inputs.cmdline, " ")}, nil
}

// unknownOption is the first option in sorted order, so the same build file
// always names the same one. A disk takes none.
func unknownOption(options map[string]string) (string, bool) {
	for _, name := range slices.Sorted(maps.Keys(options)) {
		return name, true
	}

	return "", false
}

// diskInputs are what the lines above an output say about its disk.
type diskInputs struct {
	partitions []protocol.Partition
	cmdline    []string
}

func (d *diskInputs) add(instruction imagefile.Instruction) {
	switch step := instruction.(type) {
	case imagefile.Partition:
		d.partitions = append(d.partitions, partitionOf(step))
	case imagefile.Cmdline:
		d.cmdline = append(d.cmdline, step.Text)
	}
}

func partitionOf(partition imagefile.Partition) protocol.Partition {
	settings := make([]protocol.Setting, 0, len(partition.Settings))
	for _, setting := range partition.Settings {
		settings = append(settings, protocol.Setting{Key: setting.Key, Value: setting.Value})
	}

	return protocol.Partition{Name: partition.Name, Settings: settings}
}
