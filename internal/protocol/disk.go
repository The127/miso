package protocol

// Disk asks the agent to make a bootable disk image of layers with the
// tools of another stage, and keep it as the output of the key.
type Disk struct {
	Key string

	// the keys of the image's layers, lowest first
	Layers []string

	// the keys of the layers of the stage that brings the tools, lowest
	// first
	Tools []string

	// the disk boots from an optical drive too, as an ISO
	ElTorito bool
}

func (d Disk) into(e *envelope) { e.Disk = &d }
