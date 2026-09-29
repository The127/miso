package check

import (
	"io"
	"slices"
	"strings"
)

// noticeListener takes the connections the image's systemd makes to tell
// the host how its boot goes, one for each notification, from the machine
// with the context ID.
type noticeListener interface {
	AcceptFrom(cid uint32) (io.ReadWriteCloser, error)
}

// booted waits until the image's systemd in the VM with the context ID says
// it is ready. Its notices are taken on after that until the owner of the
// notices closes them, because systemd waits on a notice nobody takes and
// then starts no check shell.
func booted(notices noticeListener, cid uint32) error {
	for {
		said, err := take(notices, cid)
		if err != nil {
			return err
		}

		// a notification is one assignment per line, and a free text such as
		// STATUS= may hold anything
		if slices.Contains(strings.Split(said, "\n"), "READY=1") {
			go func() {
				for {
					if _, err := take(notices, cid); err != nil {
						return
					}
				}
			}()

			return nil
		}
	}
}

// take reads the next notice of the VM to its end. Only a failing Accept
// ends the taking.
func take(notices noticeListener, cid uint32) (string, error) {
	conn, err := notices.AcceptFrom(cid)
	if err != nil {
		return "", err
	}

	defer func() { _ = conn.Close() }()

	return heard(conn), nil
}

// noticeSize is the most a notice may say, the PIPE_BUF up to which
// systemd itself takes a notice. More is never read, so no image has the
// host keep all it sends.
const noticeSize = 4096

// heard is what a notice said, up to where it broke. A line cut off there
// is never a whole READY=1.
func heard(notice io.Reader) string {
	said, _ := io.ReadAll(io.LimitReader(notice, noticeSize))

	return string(said)
}
