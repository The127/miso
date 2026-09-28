package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"

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

	err = a.overlaid(request.Layers, work.Dir(), func(root string) (bool, error) {
		arrived, err := placeAll(place.Open(root), request, entries)
		if err != nil {
			return false, err
		}

		if !slices.Equal([]string{arrived}, request.Digests) {
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

// placeAll puts every entry where it lands, and hands back the digest of
// what it put.
func placeAll(image *place.Root, request protocol.Copy, entries protocol.Entries) (string, error) {
	var sums []string
	for {
		entry, content, err := entries.Next()
		if errors.Is(err, io.EOF) {
			return copydigest.Of(sums), nil
		}

		if err != nil {
			return "", err
		}

		hash := sha256.New()
		if err := placeOne(image, target(request, entry), entry, io.TeeReader(content, hash)); err != nil {
			return "", err
		}

		// what was not placed is part of the entry too
		if _, err := io.Copy(hash, content); err != nil {
			return "", err
		}

		sums = append(sums, summed(entry).Sum(hex.EncodeToString(hash.Sum(nil))))
	}
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
