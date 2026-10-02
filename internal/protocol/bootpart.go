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

	// which part of the kernel to keep
	Part string
}

func (b BootPart) into(e *envelope) { e.BootPart = &b }
