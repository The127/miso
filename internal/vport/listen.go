package vport

import (
	"io"

	"github.com/hashicorp/yamux"
)

// Listener is the VM's side of a port.
type Listener struct {
	session *yamux.Session
}

// Listen opens the VM's side of the port it is handed.
func Listen(port io.ReadWriteCloser) (*Listener, error) {
	session, err := yamux.Server(port, yamux.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &Listener{session: session}, nil
}

// Accept waits for the next connection from the host.
func (l *Listener) Accept() (io.ReadWriteCloser, error) {
	return l.session.Accept()
}
