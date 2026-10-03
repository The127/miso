package plan

import (
	"slices"

	"github.com/The127/miso/internal/imagefile"
)

// The kinds of output miso makes.
const (
	KindDisk     = "disk"
	KindISO      = "iso"
	KindRootfs   = "rootfs"
	KindPortable = "portable"
	KindSysext   = "sysext"
	KindConfext  = "confext"
	KindUpdate   = "update"
	KindKernel   = "kernel"
	KindInitrd   = "initrd"
)

// The options an output takes.
const (
	// OptionELF asks for the kernel as the ELF file it unpacks to.
	OptionELF = "elf"

	// OptionFormat names the file system of a rootfs or a wrapped image.
	OptionFormat = "format"

	// OptionVersion names the version of an update.
	OptionVersion = "version"
)

// kind is what miso knows about a kind of output. Adding a kind is adding
// its constant above and its line to kinds.
type kind struct {
	// made with the tools stage. A kernel is made with it only when it is
	// unpacked.
	tools bool

	// no check can run in it
	neverBooted bool

	// a file system in a disk of its own, which is a portable, a sysext or
	// a confext
	wrapped bool

	// the options it takes
	options []string
}

var kinds = map[string]kind{
	KindDisk:     {tools: true},
	KindISO:      {tools: true},
	KindRootfs:   {tools: true, options: []string{OptionFormat}},
	KindPortable: {tools: true, neverBooted: true, wrapped: true, options: []string{OptionFormat}},
	KindSysext:   {tools: true, neverBooted: true, wrapped: true, options: []string{OptionFormat}},
	KindConfext:  {tools: true, neverBooted: true, wrapped: true, options: []string{OptionFormat}},
	KindUpdate:   {tools: true, neverBooted: true, options: []string{OptionVersion}},
	KindKernel:   {neverBooted: true, options: []string{OptionELF}},
	KindInitrd:   {neverBooted: true},
}

// Known is a kind of output miso makes.
func Known(name string) bool {
	_, known := kinds[name]

	return known
}

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
	return kinds[output.Kind].tools || Unpacks(output)
}

// NeverBooted is a kind of output that no check can run in. Unknown kinds
// are not listed, they are refused when the requests are made.
func NeverBooted(name string) bool {
	return kinds[name].neverBooted
}

// Wrapped is a kind of output that is a file system in a disk of its own.
func Wrapped(name string) bool {
	return kinds[name].wrapped
}

// Options are the options an output of a kind takes.
func Options(name string) []string {
	return kinds[name].options
}

// TakesFormat says whether an output of a kind names its file system.
func TakesFormat(name string) bool {
	return slices.Contains(Options(name), OptionFormat)
}
