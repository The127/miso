package check

import (
	"io"
	"strconv"
	"strings"
)

// Conn is the host's end of a connection to the image's check shell.
type Conn interface {
	io.ReadWriter

	// the shell reads the check until the host closes its side
	CloseWrite() error
}

// Result is what a check printed and how it exited.
type Result struct {
	Output string
	Code   int
}

// Run runs one check on its own connection.
func Run(conn Conn, command string) (Result, error) {
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
	at := strings.LastIndex(string(answer), "\n"+exitMarker+" ")

	code, err := strconv.Atoi(strings.TrimSpace(string(answer[at+len(exitMarker)+2:])))
	if err != nil {
		return Result{}, err
	}

	return Result{Output: string(answer[:at]), Code: code}, nil
}
