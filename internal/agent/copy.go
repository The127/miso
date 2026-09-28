package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"

	"github.com/The127/miso/internal/copydigest"
	"github.com/The127/miso/internal/place"
	"github.com/The127/miso/internal/protocol"
)

// ErrNotPlanned is a copy whose entries are not what the plan said of
// them. Its key would stand for other bytes than the ones planned.
var ErrNotPlanned = errors.New("not what was planned")

// ErrUnknownSource is an entry of a source that the copy does not name.
var ErrUnknownSource = errors.New("of a source the copy does not name")

// ErrOutOfOrder is an entry of a source after one of a later source. The
// sources of a copy come one after the other, in the order the build file
// names them.
var ErrOutOfOrder = errors.New("out of order")

// ErrNotLocal is an entry whose path leaves its source.
var ErrNotLocal = errors.New("leaves its source")

// ErrUnknownKind is an entry that is no file, no directory and no link. A
// copy carries nothing else.
var ErrUnknownKind = errors.New("unknown kind of entry")

// Copy puts what the host sends on top of layers and keeps it as the layer
// of the key. A key whose layer is there already needs nothing sent.
func (a *Agent) Copy(_ context.Context, request protocol.Copy, entries protocol.Entries, _ io.Writer) error {
	there, err := a.layers.Has(request.Key)
	if err != nil || there {
		return err
	}

	work, err := a.layers.Begin(request.Key)
	if err != nil {
		return err
	}

	err = a.overlaid(request.Layers, empty, work.Dir(), func(root string) (bool, error) {
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

	return work.Finish()
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

		if entry.Source < 0 || entry.Source >= len(request.Sources) {
			return nil, fmt.Errorf("%s: %w: %d", entry.Path, ErrUnknownSource, entry.Source)
		}

		if !filepath.IsLocal(entry.Path) {
			return nil, fmt.Errorf("%s: %w", entry.Path, ErrNotLocal)
		}

		// each source is checked on its own, so its place among the others is
		// checked here, or the last to write would win
		if entry.Source < source {
			return nil, fmt.Errorf("%s: %w: source %d after %d", entry.Path, ErrOutOfOrder, entry.Source, source)
		}

		source = entry.Source

		path, err := target(image, request, entry)
		if err != nil {
			return nil, err
		}

		hash := sha256.New()
		if err := placeOne(image, path, entry, io.TeeReader(content, hash)); err != nil {
			return nil, err
		}

		// what was not placed is part of the entry too
		if _, err := io.Copy(hash, content); err != nil {
			return nil, err
		}

		sums[entry.Source] = append(sums[entry.Source], summed(entry).Sum(hex.EncodeToString(hash.Sum(nil))))
	}

	digests := make([]string, 0, len(sums))
	for _, source := range sums {
		digests = append(digests, copydigest.Of(source))
	}

	return digests, nil
}

// summed is an entry as its digest sees it.
func summed(entry protocol.Entry) copydigest.Entry {
	mode := strconv.FormatUint(uint64(entry.Mode), 8)
	// a link has no mode of its own
	if entry.Kind == "link" {
		mode = ""
	}

	return copydigest.Entry{Kind: entry.Kind, Path: entry.Path, Mode: mode, Target: entry.Target}
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
