package protocol

// The parts of an image's kernel a BootPart can be.
const (
	PartKernel = "kernel"
	PartInitrd = "initrd"
)

// BootPart asks the agent to keep a part of the kernel an image boots with,
// as the output of the key.
type BootPart struct {
	Key string

	// the keys of the image's layers, lowest first
	Layers []string

	// the keys of the layers of the stage that brings the tools, lowest
	// first, for a kernel that is unpacked
	Tools []string

	// which part of the kernel to keep
	Part string

	// the kernel is unpacked into the ELF file it boots from
	ELF bool

	// the file of the image to keep, empty for the one the image installs
	Path string
}

func (b BootPart) into(e *envelope) { e.BootPart = &b }
