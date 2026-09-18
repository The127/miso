package disk

import (
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"
)

// ErrNoPartition is a number no partition of a disk has.
var ErrNoPartition = errors.New("no partition with that number")

// PartitionName names the partition of a disk in a /sys/block that has the
// number given.
func PartitionName(block fs.FS, disk string, number int64) (string, error) {
	files, err := fs.Glob(block, disk+"/*/partition")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		got, err := fs.ReadFile(block, file)
		if err != nil {
			return "", err
		}

		if strings.TrimSpace(string(got)) == strconv.FormatInt(number, 10) {
			return path.Base(path.Dir(file)), nil
		}
	}

	return "", fmt.Errorf("%s %d: %w", disk, number, ErrNoPartition)
}
