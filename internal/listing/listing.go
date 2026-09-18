package listing

import (
	"fmt"
	"io"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

// Write lists a plan, a stage as its FROM line and every step as its line
// number, its key and the instruction as written.
func Write(w io.Writer, planned plan.Plan) error {
	for _, stage := range planned.Stages {
		if _, err := fmt.Fprintf(w, "%s\n", from(stage)); err != nil {
			return err
		}

		for _, step := range stage.Steps {
			line, text := describe(step.Instruction)
			if _, err := fmt.Fprintf(w, "%4d  %s  %s\n", line, short(step.Key), text); err != nil {
				return err
			}
		}
	}

	return nil
}

func from(stage plan.Stage) string {
	line := "FROM " + stage.Base
	if stage.Name != "" {
		line += " AS " + stage.Name
	}

	if stage.BaseDigest != "" {
		line += "  base " + short(stage.BaseDigest)
	}

	return line
}

// short cuts a key to twelve characters, as git does with a commit.
func short(key string) string {
	return key[:12]
}

// describe is an instruction as its author wrote it, and where.
func describe(instruction imagefile.Instruction) (int, string) {
	switch step := instruction.(type) {
	case imagefile.Run:
		return step.Line, "RUN " + step.Command
	case imagefile.Env:
		return step.Line, "ENV " + step.Key + "=" + step.Value
	}

	return 0, ""
}
