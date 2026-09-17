package imagefile

import (
	"errors"
	"fmt"
)

var (
	// ErrUnknownInstruction is a line that starts with no keyword we know.
	ErrUnknownInstruction = errors.New("unknown instruction")

	// ErrBeforeFrom is an instruction with no stage to belong to.
	ErrBeforeFrom = errors.New("before FROM")
)

// Error is a problem in a build file, at a line.
type Error struct {
	Line int
	Err  error
}

func (e *Error) Error() string {
	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}
