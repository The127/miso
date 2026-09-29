package agent

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/The127/miso/internal/copydigest"
	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
)

// ErrNotPlanned is a copy whose entries are not what the plan said of
// them. Its key would stand for other bytes than the ones planned.
var ErrNotPlanned = errors.New("not what was planned")

// ErrUnknownKind is an entry that is no file, no directory and no link. A
// copy carries nothing else.
var ErrUnknownKind = errors.New("unknown kind of entry")

// Copy puts what the host sends, or what an earlier stage's layers hold,
// on top of layers and keeps it as the layer of the key. A key whose layer
// is there already needs nothing sent.
func (a *Agent) Copy(_ context.Context, request protocol.Copy, entries protocol.Entries, _ io.Writer) error {
	there, at, err := a.found(request.Key, slices.Concat(request.Layers, request.From))
	if err != nil || there {
		return err
	}

	work, err := a.layers.Begin(request.Key)
	if err != nil {
		return err
	}

	err = a.overlaid(request.Layers, empty, nil, work.Dir(), func(root string) (bool, error) {
		// a stage's files are in its layers, and the plan's key covers them
		if request.Stage != "" {
			err := a.copyStage(request, root)

			return err == nil, err
		}

		arrived, err := placeAll(place.Open(root), request, entries)
		if err != nil {
			return false, err
		}

		if !slices.Equal(arrived, request.Digests) {
			return false, fmt.Errorf("%s: %w", request.Destination, ErrNotPlanned)
		}

		return true, nil
	})
	if err != nil {
		_ = work.Discard()

		return err
	}

	return work.FinishAt(at)
}

// empty is the bottom of a copy, with nothing of a run's in it, so that
// what the copy puts is all its layer holds. Open to all, as the root of
// an image on scratch is.
func empty(scratch string) (string, error) {
	dir := filepath.Join(scratch, "empty")
	if err := os.Mkdir(dir, 0o700); err != nil {
		return "", err
	}

	return dir, os.Chmod(dir, 0o755) //nolint:gosec // an image's root is open to all
}

// placeAll puts every entry where it lands, and hands back the digest of
// what it put of each source.
func placeAll(image *place.Root, request protocol.Copy, entries protocol.Entries) ([]string, error) {
	sums := make([][]string, len(request.Sources))
	source := 0
	for {
		entry, content, err := entries.Next()
		if errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return nil, err
		}

		if err := checked(request, entry, source); err != nil {
			return nil, err
		}

		source = entry.Source

		path, err := target(image, request, entry)
		if err != nil {
			return nil, err
		}

		passing := copydigest.Pass(content)
		if err := placeOne(image, path, entry, passing); err != nil {
			return nil, err
		}

		sum, err := passing.Sum(summed(entry))
		if err != nil {
			return nil, err
		}

		sums[entry.Source] = append(sums[entry.Source], sum)
	}

	digests := make([]string, 0, len(sums))
	for _, source := range sums {
		digests = append(digests, copydigest.Of(source))
	}

	return digests, nil
}

// summed is an entry as its digest sees it.
func summed(entry protocol.Entry) copydigest.Entry {
	return copydigest.Entry{Kind: entry.Kind, Path: entry.Path, Mode: entry.Mode, Target: entry.Target}
}

// placeOne puts an entry at a path of the image as what it is.
func placeOne(image *place.Root, path string, entry protocol.Entry, content io.Reader) error {
	switch entry.Kind {
	case "directory":
		return image.Directory(path, entry.Mode)
	case "link":
		return image.Link(path, entry.Target)
	case "file":
		return image.File(path, entry.Mode, content)
	default:
		return fmt.Errorf("%s: %w: %s", path, ErrUnknownKind, entry.Kind)
	}
}
