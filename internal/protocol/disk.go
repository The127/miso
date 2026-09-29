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

	// the partitions of the disk, in the order they are laid out
	Partitions []Partition
}

// Partition is a definition for systemd-repart, named as its file is.
type Partition struct {
	Name     string
	Settings []Setting
}

// Setting is a line of a definition. A key may come more than once.
type Setting struct {
	Key   string
	Value string
}

func (d Disk) into(e *envelope) { e.Disk = &d }
