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
			if _, err := fmt.Fprintf(w, "%4d  %s  %s\n", line(step.Instruction), short(step.Key), text(step.Instruction)); err != nil {
				return err
			}
		}
	}

	return nil
}

func from(stage plan.Stage) string {
	if stage.Name != "" {
		return "FROM " + stage.Base + " AS " + stage.Name
	}

	return "FROM " + stage.Base
}

// short cuts a key to twelve characters, as git does with a commit.
func short(key string) string {
	return key[:12]
}

func line(instruction imagefile.Instruction) int {
	if run, isRun := instruction.(imagefile.Run); isRun {
		return run.Line
	}

	return 0
}

func text(instruction imagefile.Instruction) string {
	if run, isRun := instruction.(imagefile.Run); isRun {
		return "RUN " + run.Command
	}

	return ""
}
