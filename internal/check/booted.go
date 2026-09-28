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

// Booted waits until the image's systemd says it is ready. Its notices are
// taken on after that until the owner of the notices closes them, because
// systemd waits on a notice nobody takes and then starts no check shell.
func Booted(notices Notices) error {
	for {
		said, err := take(notices)
		if err != nil {
			return err
		}

		// a notification is one assignment per line, and a free text such as
		// STATUS= may hold anything
		if slices.Contains(strings.Split(said, "\n"), "READY=1") {
			go func() {
				for {
					if _, err := take(notices); err != nil {
						return
					}
				}
			}()

			return nil
		}
	}
}

// take reads the next notice to its end. Only a failing Accept ends the
// taking.
func take(notices Notices) (string, error) {
	conn, err := notices.Accept()
	if err != nil {
		return "", err
	}

	defer func() { _ = conn.Close() }()

	return heard(conn), nil
}

// heard is what a notice said, or nothing for one that broke, since what it
// said is not known.
func heard(notice io.Reader) string {
	said, err := io.ReadAll(notice)
	if err != nil {
		return ""
	}

	return string(said)
}
