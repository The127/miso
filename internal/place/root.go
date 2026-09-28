package place

// Root is the root file system of an image, a directory of the builder VM.
type Root struct {
	dir string
}

// Open takes the directory an image's root file system is mounted on. It
// touches nothing yet.
func Open(dir string) *Root {
	return &Root{dir: dir}
}
