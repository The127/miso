package main

import (
	"fmt"
	"math"

	"github.com/dustin/go-humanize"
)

// sizeOf is a size in bytes that the text of a flag says, like 10GiB.
func sizeOf(flag, text string) (int64, error) {
	size, err := humanize.ParseBytes(text)
	if err != nil {
		return 0, fmt.Errorf("%s %q is not a size like 10GiB: %w", flag, text, err)
	}

	// a wrapped size would be negative
	if size > math.MaxInt64 {
		return 0, fmt.Errorf("%s %q is too large", flag, text)
	}

	if size == 0 {
		return 0, fmt.Errorf("%s %q must be above zero", flag, text)
	}

	return int64(size), nil
}

// ibytes is a size of a disk, which is never negative, as humanize writes it.
func ibytes(size int64) string {
	return humanize.IBytes(uint64(size)) //nolint:gosec // the size of a file or of a disk
}
