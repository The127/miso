package protocol

import "io"

// Run asks the agent to run a command and writes what the command writes
// to out until the agent tells how it ended.
func (c *Conn) Run(run Run, out io.Writer) error {
	if err := c.Send(run); err != nil {
		return err
	}

	for {
		message, err := c.Receive()
		if err != nil {
			return err
		}

		switch m := message.(type) {
		case Output:
			if _, err := out.Write(m.Bytes); err != nil {
				return err
			}
		case Done:
			return nil
		}
	}
}
