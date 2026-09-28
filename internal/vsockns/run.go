package vsockns

import (
	"encoding/json"
	"os"
)

// Run runs a program inside the namespace to its end, writing to stdout.
func (n *Namespace) Run(args []string, stdout *os.File) error {
	argument, err := json.Marshal(args)
	if err != nil {
		return err
	}

	_, _, err = n.ask(runQuestion+" "+string(argument), int(stdout.Fd()))

	return err
}
