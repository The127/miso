package plan

import (
	"slices"

	"github.com/The127/miso/internal/imagefile"
)

// The kinds of output miso makes.
const (
	KindDisk   = "disk"
	KindISO    = "iso"
	KindRootfs = "rootfs"
	KindKernel = "kernel"
	KindInitrd = "initrd"
)

// Known is a kind of output miso makes.
func Known(kind string) bool {
	return slices.Contains([]string{KindDisk, KindISO, KindRootfs, KindKernel, KindInitrd}, kind)
}

// OptionELF asks for the kernel as the ELF file it unpacks to.
const OptionELF = "elf"

// Unpacks is an output that asks for the ELF file of a kernel. A bare
// --elf and --elf= say the same.
func Unpacks(output imagefile.Output) bool {
	_, asked := output.Options[OptionELF]

	return output.Kind == KindKernel && asked
}

// NeedsTools says whether an output is made with the tools stage. The kernel
// and the initrd are files of the image, copied out as they are, unless the
// kernel is unpacked.
func NeedsTools(output imagefile.Output) bool {
	if output.Kind == KindKernel {
		return Unpacks(output)
	}

	return slices.Contains([]string{KindDisk, KindISO, KindRootfs}, output.Kind)
}

// NeverBooted is a kind of output that no check can run in. Unknown kinds
// are not listed, they are refused when the requests are made.
func NeverBooted(kind string) bool {
	return slices.Contains([]string{KindRootfs, KindKernel, KindInitrd}, kind)
}
