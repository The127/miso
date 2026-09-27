package builder

import "io"

// Dial opens a new connection to the agent.
type Dial func() (io.ReadWriteCloser, error)
