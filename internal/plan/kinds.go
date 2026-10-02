package plan

import "slices"

// The kinds of output miso makes.
const (
	KindDisk   = "disk"
	KindISO    = "iso"
	KindRootfs = "rootfs"
)

// Known is a kind of output miso makes.
func Known(kind string) bool {
	return slices.Contains([]string{KindDisk, KindISO, KindRootfs}, kind)
}
