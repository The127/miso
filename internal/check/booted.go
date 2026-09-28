package check

import (
	"bytes"
	"io"
)

// Notices are the connections the image's systemd makes to tell the host
// how its boot goes, one for each notification.
type Notices interface {
	Accept() (io.ReadWriteCloser, error)
}

// Booted waits until the image's systemd says it is ready.
func Booted(notices Notices) error {
	for {
		conn, err := notices.Accept()
		if err != nil {
			return err
		}

		said, err := io.ReadAll(conn)
		_ = conn.Close()

		if err != nil {
			return err
		}

		if bytes.Contains(said, []byte("READY=1")) {
			return nil
		}
	}
}
