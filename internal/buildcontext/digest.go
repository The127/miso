package buildcontext

import (
	"io"
	"strings"

	"github.com/The127/miso/internal/copydigest"
)

// Digest says what a COPY of this path would put into an image.
func (d *Dir) Digest(path string) (string, error) {
	var sums []string
	err := d.walk(path, func(found entry) error {
		content, _, err := d.contentOf(found)
		if err != nil {
			return err
		}

		defer func() { _ = content.Close() }()

		sum, err := copydigest.Pass(content).Sum(found.digested())
		sums = append(sums, sum)

		return err
	})
	if err != nil {
		return "", err
	}

	return copydigest.Of(sums), nil
}

// digested is the entry as its digest sees it.
func (e entry) digested() copydigest.Entry {
	return copydigest.Entry{Kind: string(e.kind), Path: e.path, Mode: e.mode, Target: e.target}
}

// contentOf is what an entry carries, and how much of it. Only a file
// carries anything.
func (d *Dir) contentOf(found entry) (io.ReadCloser, int64, error) {
	if found.kind != kindFile {
		return io.NopCloser(strings.NewReader("")), 0, nil
	}

	file, err := d.open(found.name, found.looked)
	if err != nil {
		return nil, 0, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()

		return nil, 0, err
	}

	return file, info.Size(), nil
}
