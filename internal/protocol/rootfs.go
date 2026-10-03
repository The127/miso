package protocol

// The file systems a Rootfs can be.
const (
	FormatExt4  = "ext4"
	FormatErofs = "erofs"
)

// The images a file system can be wrapped in.
const (
	WrapPortable = "portable"
	WrapSysext   = "sysext"
	WrapConfext  = "confext"
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

	// the name of the output file, which an extension is named by
	Name string

	// the kind of image the file system is wrapped in, a disk of its own,
	// empty when it is not wrapped
	Wrap string
}

func (r Rootfs) into(e *envelope) { e.Rootfs = &r }
