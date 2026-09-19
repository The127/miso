package initramfs

// Module is a kernel module, unpacked, that the agent loads at boot.
type Module struct {
	Name    string
	Content []byte
}
