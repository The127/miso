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

// outputRequest is what an output asks for on the root file system under it.
// A disk, an ISO or a rootfs takes the tools from where their stage ends, the
// kernel and the initrd are taken from the image, and an unpacked kernel
// takes the tools too.
func outputRequest(step plan.Step, instruction imagefile.Output, under rootfs, tools rootfs, inputs diskInputs) (protocol.Message, error) {
	if refused := refusal(instruction); refused != nil {
		line, written := imagefile.Written(instruction)

		return nil, &imagefile.Error{Line: line, Err: fmt.Errorf("%s: %w", written, refused)}
	}

	if part, isPart := bootParts[instruction.Kind]; isPart {
		return protocol.BootPart{Key: step.Key, Layers: under.layers, Tools: tools.layers, Part: part, ELF: plan.Unpacks(instruction)}, nil
	}

	if takesFormat(instruction.Kind) {
		return protocol.Rootfs{Key: step.Key, Layers: under.layers, Tools: tools.layers, Format: formatOf(instruction.Options), Wrap: wrapOf(instruction.Kind)}, nil
	}

	return protocol.Disk{Key: step.Key, Layers: under.layers, Tools: tools.layers, ElTorito: instruction.Kind == plan.KindISO, Partitions: slices.Clone(inputs.partitions), Cmdline: strings.Join(inputs.cmdline, " ")}, nil
}

// refusal is why miso cannot make an output, or nil when it can.
func refusal(instruction imagefile.Output) error {
	unknown, hasUnknown := unknownOption(instruction.Options, optionsOf(instruction.Kind))
	switch {
	case !plan.Known(instruction.Kind):
		return ErrUnknownKind
	case hasUnknown:
		return fmt.Errorf("--%s: %w", unknown, ErrUnknownOption)
	case instruction.Options[plan.OptionELF] != "":
		return fmt.Errorf("--%s: %w", plan.OptionELF, ErrOptionTakesNoValue)
	case takesFormat(instruction.Kind) && !slices.Contains(formats, formatOf(instruction.Options)):
		return fmt.Errorf("--format=%s: %w", formatOf(instruction.Options), ErrUnknownFormat)
	}

	return nil
}

// wrapOf is the image a file system of a kind is wrapped in, empty when it is
// not.
func wrapOf(kind string) string {
	if kind == plan.KindPortable {
		return protocol.WrapPortable
	}

	return ""
}

// takesFormat says whether an output of a kind is a file system of a format.
func takesFormat(kind string) bool {
	return kind == plan.KindRootfs || kind == plan.KindPortable
}

// formatOption names the file system of a rootfs.
const formatOption = "format"

// formats are the file systems a rootfs can be.
var formats = []string{protocol.FormatExt4, protocol.FormatErofs}

// formatOf is the file system an output asks for, ext4 when it names none.
func formatOf(options map[string]string) string {
	format, given := options[formatOption]
	if !given {
		return protocol.FormatExt4
	}

	return format
}

// optionsOf are the options an output of a kind takes.
func optionsOf(kind string) []string {
	switch kind {
	case plan.KindRootfs, plan.KindPortable:
		return []string{formatOption}
	case plan.KindKernel:
		return []string{plan.OptionELF}
	}

	return nil
}

// unknownOption is the first option in sorted order that the output does not
// take, so the same build file always names the same one.
func unknownOption(options map[string]string, taken []string) (string, bool) {
	for _, name := range slices.Sorted(maps.Keys(options)) {
		if !slices.Contains(taken, name) {
			return name, true
		}
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
