package vport

import (
	"io"

	"github.com/hashicorp/yamux"
)

// Dialer is the host's side of a port.
type Dialer struct {
	session *yamux.Session
}

// Connect opens the host's side of the port it is handed.
func Connect(port io.ReadWriteCloser) (*Dialer, error) {
	session, err := yamux.Client(port, yamux.DefaultConfig())
	if err != nil {
		return nil, err
	}

	return &Dialer{session: session}, nil
}

// Dial opens a connection to the VM.
func (d *Dialer) Dial() (io.ReadWriteCloser, error) {
	return d.session.Open()
}
