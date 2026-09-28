package check

import (
	"io"
	"slices"
	"strings"
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

		// a notification is one assignment per line, and a free text such as
		// STATUS= may hold anything
		if slices.Contains(strings.Split(string(said), "\n"), "READY=1") {
			return nil
		}
	}
}
