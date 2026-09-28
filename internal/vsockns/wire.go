package vsockns

import (
	"io"

	"golang.org/x/sys/unix"
)

// messageBytes is the most a message between miso and its helper holds.
const messageBytes = 4096

// say sends a message, with the files it hands over.
func say(conn int, text string, files ...int) error {
	var rights []byte
	if len(files) > 0 {
		rights = unix.UnixRights(files...)
	}

	return unix.Sendmsg(conn, []byte(text), rights, nil, 0)
}

// hear takes the next message and the files that came with it.
func hear(conn int) (string, []int, error) {
	text := make([]byte, messageBytes)
	oob := make([]byte, unix.CmsgSpace(4))

	n, oobn, _, _, err := unix.Recvmsg(conn, text, oob, unix.MSG_CMSG_CLOEXEC)
	if err != nil {
		return "", nil, err
	}

	// the other side is gone
	if n == 0 && oobn == 0 {
		return "", nil, io.EOF
	}

	var files []int
	messages, err := unix.ParseSocketControlMessage(oob[:oobn])
	if err != nil {
		return "", nil, err
	}

	for _, message := range messages {
		fds, err := unix.ParseUnixRights(&message)
		if err != nil {
			return "", nil, err
		}

		files = append(files, fds...)
	}

	return string(text[:n]), files, nil
}
