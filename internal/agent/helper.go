package agent

import (
	"fmt"
	"os"
	"syscall"
)

// helperName is the name the agent starts itself under to run a command.
const helperName = "miso-run"

// Helper sets up a run and becomes its shell when the agent was started as
// the helper of a run, and returns at once otherwise.
func Helper() {
	if os.Args[0] != helperName {
		return
	}

	// the agent reads why a run could not start from here, and the start of
	// the shell closes it
	failed := os.NewFile(3, "failed")
	syscall.CloseOnExec(3)
	root, command := os.Args[1], os.Args[2]
	// returns only when the shell did not start
	err := helper(root, command)
	_, _ = fmt.Fprint(failed, err)
	os.Exit(1)
}

func helper(root, command string) error {
	if err := enterRoot(root); err != nil {
		return err
	}

	for _, m := range mounts {
		if err := m.mount(); err != nil {
			return err
		}
	}

	if err := upLoopback(); err != nil {
		return err
	}

	if err := nameRun(); err != nil {
		return err
	}

	err := syscall.Exec("/bin/sh", []string{"/bin/sh", "-c", command}, os.Environ()) //nolint:gosec // running what the build file says is what a RUN is

	return fmt.Errorf("run /bin/sh: %w", err)
}
