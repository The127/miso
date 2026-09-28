package qemu

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// agentPort is the name the agent finds its virtio port by.
const agentPort = "miso"

// port gives the machine a virtio serial port for its agent, on the socket
// QEMU finds at fd.
func port(fd int) []string {
	return []string{
		"-device", "virtio-serial-pci",
		"-chardev", fmt.Sprintf("socket,id=agent,fd=%d", fd),
		"-device", "virtserialport,chardev=agent,name=" + agentPort,
	}
}

// socketPair is a connected pair of sockets, one end for the host and one
// for QEMU, so the port needs no file of its own.
func socketPair() (host *os.File, machine *os.File, err error) {
	fds, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return nil, nil, err
	}

	return os.NewFile(uintptr(fds[0]), "agent port"), os.NewFile(uintptr(fds[1]), "agent port of QEMU"), nil
}
