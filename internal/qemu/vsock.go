package qemu

import "fmt"

// vsock gives the machine the CID the host claimed on the open vhost-vsock
// device it hands QEMU, so no other machine can take that CID in between.
func vsock(cid uint32, fd int) []string {
	return []string{"-device", fmt.Sprintf("vhost-vsock-pci,guest-cid=%d,vhostfd=%d", cid, fd)}
}
