package imagefile

import (
	"errors"
	"fmt"
)

var (
	// ErrUnknownInstruction is a line that starts with no keyword we know.
	ErrUnknownInstruction = errors.New("unknown instruction")

	// ErrNoFrom is a build file without a single stage.
	ErrNoFrom = errors.New("no FROM instruction")

	// ErrBeforeFrom is an instruction with no stage to belong to.
	ErrBeforeFrom = errors.New("before FROM")

	// ErrContinuation is a backslash with no line to continue on.
	ErrContinuation = errors.New("continuation")

	// ErrArguments is a known instruction with arguments it cannot take.
	ErrArguments = errors.New("bad arguments")
)

// Error is a problem in a build file, at a line.
type Error struct {
	File string
	Line int
	Err  error
}

func (e *Error) Error() string {
	if e.File != "" {
		return fmt.Sprintf("%s:%d: %v", e.File, e.Line, e.Err)
	}

	return fmt.Sprintf("line %d: %v", e.Line, e.Err)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// InFile names the file of the line an error is at.
func InFile(file string, err error) error {
	var located *Error
	if !errors.As(err, &located) {
		return err
	}

	named := *located
	named.File = file

	return &named
}

// argumentsError is what a reader objects to, said after its keyword.
type argumentsError struct {
	keyword   string
	complaint error
}

func (e *argumentsError) Error() string {
	return e.keyword + " " + e.complaint.Error()
}

func (e *argumentsError) Is(target error) bool {
	return target == ErrArguments
}
