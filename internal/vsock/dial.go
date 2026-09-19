package vsock

import mdvsock "github.com/mdlayher/vsock"

// Dial connects to a port of the machine with a context ID.
func Dial(cid, port uint32) (*mdvsock.Conn, error) {
	return mdvsock.Dial(cid, port, nil)
}
