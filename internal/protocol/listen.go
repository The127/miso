package protocol

import "io"

// Listener hands the agent the connections of the host.
type Listener interface {
	Accept() (io.ReadWriteCloser, error)
}

// Serve answers the host on every connection the listener accepts, until
// accepting fails.
func Serve(listener Listener, agent string, runner Runner) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		_ = New(agent, conn, conn).Serve(runner)
	}
}
