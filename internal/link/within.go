package link

import (
	"fmt"
	"os"
	"runtime"

	"golang.org/x/sys/unix"
)

// Within runs a call on a thread in a network namespace, or in the
// caller's own one for nil. What the call opens stays in that namespace.
func Within(namespace *os.File, call func() error) error {
	done := make(chan error)
	go func() {
		if namespace != nil {
			// never unlocked, so the thread that moved ends with this goroutine
			runtime.LockOSThread()
			if err := unix.Setns(int(namespace.Fd()), unix.CLONE_NEWNET); err != nil {
				done <- fmt.Errorf("enter the network: %w", err)

				return
			}
		}

		done <- call()
	}()

	return <-done
}
