package disk

import (
	"io/fs"
	"path"
)

// BySerial names the disk in a /sys/block whose serial is the one given.
func BySerial(block fs.FS, serial string) (string, error) {
	files, err := fs.Glob(block, "*/serial")
	if err != nil {
		return "", err
	}

	for _, file := range files {
		got, err := fs.ReadFile(block, file)
		if err != nil {
			return "", err
		}

		if string(got) == serial {
			return path.Dir(file), nil
		}
	}

	return "", nil
}
