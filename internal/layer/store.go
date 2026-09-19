package layer

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
)

// what is not a finished layer is named with these, and a key is hex, so no
// key is ever named like them
const (
	working = "work-"
	scratch = "scratch-"
)

// Store keeps the layers of one cache directory, each under its key.
type Store struct {
	dir string
}

// Open takes the directory the layers live in. It touches nothing yet.
func Open(dir string) *Store {
	return &Store{dir: dir}
}

// Path is where the layer of a key lives.
func (s *Store) Path(key string) string {
	return filepath.Join(s.dir, key)
}

// Has tells whether the layer of a key is finished.
func (s *Store) Has(key string) (bool, error) {
	_, err := os.Stat(s.Path(key))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}

	return err == nil, err
}

// Begin starts the layer of a key in a directory of its own.
func (s *Store) Begin(key string) (*Work, error) {
	dir, err := os.MkdirTemp(s.dir, working)
	if err != nil {
		return nil, err
	}

	return &Work{dir: dir, final: s.Path(key)}, nil
}

// Scratch is a new directory on the file system of the layers, for the
// caller to remove.
func (s *Store) Scratch() (string, error) {
	return os.MkdirTemp(s.dir, scratch)
}

// Work is a layer being built.
type Work struct {
	dir   string
	final string
}

// Dir is where the files of the layer go while it is built.
func (w *Work) Dir() string {
	return w.dir
}

// Discard throws away a layer that will not be finished.
func (w *Work) Discard() error {
	return os.RemoveAll(w.dir)
}

// Finish puts the layer under its key.
func (w *Work) Finish() error {
	// a crash may lose a finished layer but must never leave a key whose
	// files the disk does not hold yet
	if err := syncFileSystem(w.dir); err != nil {
		return err
	}

	err := os.Rename(w.dir, w.final)
	// the same key is the same layer, so the one there already is as good
	if errors.Is(err, fs.ErrExist) {
		return os.RemoveAll(w.dir)
	}

	if err != nil {
		return err
	}

	// the VM is stopped by killing it, which must not take the last layer
	// of a build with it
	return syncFileSystem(w.final)
}
