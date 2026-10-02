package protocol

// The file systems a Rootfs can be.
const (
	FormatExt4  = "ext4"
	FormatErofs = "erofs"
)

// Rootfs asks the agent to make a file system image of layers with the
// tools of another stage, and keep it as the output of the key.
type Rootfs struct {
	Key string

	// the keys of the image's layers, lowest first
	Layers []string

	// the keys of the layers of the stage that brings the tools, lowest
	// first
	Tools []string

	// the file system to make
	Format string
}

func (r Rootfs) into(e *envelope) { e.Rootfs = &r }
