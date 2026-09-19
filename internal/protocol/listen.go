package protocol

import (
	"io"
	"sync"
)

// Listener hands the agent the connections of the host.
type Listener interface {
	Accept() (io.ReadWriteCloser, error)
}

// Serve answers the host on every connection the listener accepts, until
// accepting fails. It returns once every answer is sent.
func Serve(listener Listener, agent string, runner Runner) error {
	var serving sync.WaitGroup
	defer serving.Wait()

	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}

		serving.Go(func() {
			_ = New(agent, conn, conn).Serve(runner)
		})
	}
}
