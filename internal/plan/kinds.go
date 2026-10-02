package plan

import "slices"

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

// NeedsTools is a kind of output made with the tools stage. The kernel and
// the initrd are files of the image, copied out as they are.
func NeedsTools(kind string) bool {
	return slices.Contains([]string{KindDisk, KindISO, KindRootfs}, kind)
}

// NeverBooted is a kind of output that no check can run in. Unknown kinds
// are not listed, they are refused when the requests are made.
func NeverBooted(kind string) bool {
	return slices.Contains([]string{KindRootfs, KindKernel, KindInitrd}, kind)
}
