package listing

import (
	"fmt"
	"io"
	"strings"

	"github.com/The127/miso/internal/plan"
)

// Write lists a plan, a stage as its FROM line and every step as its line
// number, its key and the instruction as written.
func Write(w io.Writer, planned plan.Plan) error {
	out := &lines{w: w}
	if planned.Agent != "" {
		out.print("agent %s", planned.Agent)
	}

	for _, stage := range planned.Stages {
		out.print("%s", from(stage))
		for _, step := range stage.Steps {
			out.step(step)
		}
	}

	return out.err
}

// lines keeps the first error a writer gives, so that the listing can be
// written as its lines and the error is checked once at the end.
type lines struct {
	w   io.Writer
	err error
}

func (l *lines) print(format string, args ...any) {
	if l.err != nil {
		return
	}

	_, l.err = fmt.Fprintf(l.w, format+"\n", args...)
}

func (l *lines) step(step plan.Step) {
	line, text := describe(step.Instruction)
	l.print("%4d  %s  %s", line, short(step.Key), text)

	for _, file := range step.Files {
		l.print("%20s%s  %s", "", file.Path, short(file.Digest))
	}
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

// short cuts a key to twelve characters, as git does with a commit. A
// digest names its format in front, and that stays. A step without a key
// keeps its column.
func short(key string) string {
	if key == "" {
		return strings.Repeat("-", 12)
	}

	name, hex, named := strings.Cut(key, ":")
	if !named {
		return key[:12]
	}

	return name + ":" + hex[:12]
}
