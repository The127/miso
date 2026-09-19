package cachedisk

import (
	"errors"
	"io"
	"os"

	"golang.org/x/sys/unix"
)

// Lock holds the cache for one build until it is closed, because two
// builders writing one disk would break it. Where another build holds it,
// Lock calls waiting and then waits for it.
func Lock(path string, waiting func()) (io.Closer, error) {
	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}

	err = unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if errors.Is(err, unix.EWOULDBLOCK) {
		waiting()

		err = unix.Flock(int(file.Fd()), unix.LOCK_EX)
	}

	if err != nil {
		return nil, errors.Join(err, file.Close())
	}

	return file, nil
}
