package check

import "io"

// Conn is the host's end of a connection to the image's check shell.
type Conn interface {
	io.Writer

	// the shell reads the check until the host closes its side
	CloseWrite() error
}

// Run runs one check on its own connection.
func Run(conn Conn, command string) error {
	if _, err := io.WriteString(conn, command+"\n"); err != nil {
		return err
	}

	return conn.CloseWrite()
}
