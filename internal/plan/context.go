package plan

import (
	"fmt"

	"github.com/The127/miso/internal/imagefile"
)

// Context is the directory a COPY takes its files from.
type Context interface {
	Digest(path string) (string, error)
}

func contextFiles(step imagefile.Copy, context Context) ([]File, error) {
	var files []File
	for _, source := range step.Sources {
		digest, err := context.Digest(source)
		if err != nil {
			return nil, at(step.Line, fmt.Errorf("COPY %s: %w", source, err))
		}

		files = append(files, File{Path: source, Digest: digest})
	}

	return files, nil
}
