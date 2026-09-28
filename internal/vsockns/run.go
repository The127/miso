package vsockns

import (
	"fmt"
	"os"
)

// Run runs a program inside the namespace to its end, writing to stdout.
func (n *Namespace) Run(args []string, stdout *os.File) error {
	program, err := n.Start(args, stdout)
	if err != nil {
		return err
	}

	code, err := program.Wait()
	if err != nil {
		return err
	}

	if code != 0 {
		return fmt.Errorf("%s exited with %d", args[0], code)
	}

	return nil
}
