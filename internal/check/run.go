package check

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

// shellConn is the host's end of a connection to the image's check shell.
type shellConn interface {
	io.ReadWriter

	// CloseWrite the shell reads the check until the host closes its side
	CloseWrite() error
}

// Result is what a check printed and how it exited.
type Result struct {
	Output string
	Code   int
}

// run runs one check on its own connection.
func run(conn shellConn, command string) (Result, error) {
	if _, err := io.WriteString(conn, command+"\n"); err != nil {
		return Result{}, err
	}

	if err := conn.CloseWrite(); err != nil {
		return Result{}, err
	}

	answer, err := io.ReadAll(conn)
	if err != nil {
		return Result{}, err
	}

	// the service's echo ends the output with the line break before the marker
	markerLine := "\n" + exitMarker + " "
	at := strings.LastIndex(string(answer), markerLine)
	if at == -1 {
		return Result{}, errors.New("the check's answer ended without an exit code")
	}

	code, err := strconv.Atoi(strings.TrimSpace(string(answer[at+len(markerLine):])))
	if err != nil {
		return Result{}, err
	}

	return Result{Output: string(answer[:at]), Code: code}, nil
}
