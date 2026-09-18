package disk

import (
	"io/fs"
	"path"
	"strconv"
	"strings"
)

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

	return "", nil
}
