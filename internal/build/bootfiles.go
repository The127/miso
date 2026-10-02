package build

import (
	"strings"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

// BootFiles are what a rootfs has to be booted with: the names of the files
// that the kernel and the initrd outputs above it are fetched into, and the
// kernel command line.
type BootFiles struct {
	Kernel  string
	Initrd  string
	Cmdline string
}

// bootingOutputs are the outputs of a stage that a rootfs is booted with,
// and the requests that fetch them. The plan puts both above a rootfs that
// has a CHECK. The command line is the one at the last output, so lines
// after a rootfs do not count.
type bootingOutputs struct {
	kernel, initrd           string
	kernelFetch, initrdFetch int
	cmdline                  string
}

// newBootingOutputs has no fetch yet. The indexes start at -1 so that a
// rootfs without a kernel or an initrd above it fails loudly and never marks
// another request.
func newBootingOutputs() bootingOutputs {
	return bootingOutputs{kernelFetch: -1, initrdFetch: -1}
}

// add takes note of an output fetched by the request at an index, with the
// lines of CMDLINE above it.
func (b *bootingOutputs) add(output imagefile.Output, fetch int, cmdline []string) {
	b.cmdline = strings.Join(cmdline, " ")

	switch output.Kind {
	case plan.KindKernel:
		b.kernel, b.kernelFetch = output.Name, fetch
	case plan.KindInitrd:
		b.initrd, b.initrdFetch = output.Name, fetch
	}
}

// bootFiles marks the fetches of the kernel and the initrd as needed, and
// answers what the rootfs fetched last is booted with.
func (b *bootingOutputs) bootFiles(requests []Request) *BootFiles {
	requests[b.kernelFetch].Needed = true
	requests[b.initrdFetch].Needed = true

	return &BootFiles{Kernel: b.kernel, Initrd: b.initrd, Cmdline: b.cmdline}
}
