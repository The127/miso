package agent

import (
	"fmt"
	"syscall"
)

// nameRun gives the run its host name.
func nameRun() error {
	// a fixed name, because packages write it into what they install, and
	// every hosts file knows this one
	if err := syscall.Sethostname([]byte("localhost")); err != nil {
		return fmt.Errorf("name the run: %w", err)
	}

	return nil
}
