package listing

import (
	"fmt"
	"io"
	"strings"

	"github.com/The127/miso/internal/imagefile"
	"github.com/The127/miso/internal/plan"
)

// Write lists a plan, a stage as its FROM line and every step as its line
// number, its key and the instruction as written. A step whose key is in
// cached says so. A nil cached says nothing about the cache.
func Write(w io.Writer, planned plan.Plan, cached map[string]bool) error {
	out := &lines{w: w, cached: cached}
	if planned.Agent != "" {
		out.print("agent %s", planned.Agent)
	}

	for _, base := range planned.Downloads {
		out.print("download %s", base)
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
	w      io.Writer
	cached map[string]bool
	err    error
}

func (l *lines) print(format string, args ...any) {
	if l.err != nil {
		return
	}

	_, l.err = fmt.Fprintf(l.w, format+"\n", args...)
}

func (l *lines) step(step plan.Step) {
	line, text := imagefile.Written(step.Instruction)
	marker := l.marker(step)
	l.print("%4d  %s  %s%s", line, short(step.Key), marker, text)

	for _, file := range step.Files {
		l.print("%*s%s  %s", 20+len(marker), "", file.Path, short(file.Digest))
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

// The markers of the cache column, of one width so the text lines up.
const (
	markerCached = "cached  "
	markerRun    = "run     "
	markerNone   = "        "
)

func (l *lines) marker(step plan.Step) string {
	switch {
	case l.cached == nil:
		return ""
	case !makesLayer(step.Instruction):
		return markerNone
	case l.cached[step.Key]:
		return markerCached
	}

	return markerRun
}

// makesLayer is an instruction whose key is the key of a layer in the cache.
func makesLayer(instruction imagefile.Instruction) bool {
	switch instruction.(type) {
	case imagefile.Run, imagefile.Copy, imagefile.Output:
		return true
	}

	return false
}
