package qemu

// claim takes a CID for the machine on the host's vhost-vsock device.
func claim(take func(cid uint32) error, random func() uint32) (uint32, error) {
	cid := random()

	return cid, take(cid)
}
