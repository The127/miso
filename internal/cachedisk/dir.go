package cachedisk

import "path/filepath"

// Dir is the directory the builder keeps its cache in: the disk, the lock
// that gives it to one build at a time, and the logs of the builder VM. It
// names the files so that no caller has to.
type Dir string

// Disk is the cache disk.
func (d Dir) Disk() string { return d.file("layers.img") }

// LockFile is the file Lock is taken on.
func (d Dir) LockFile() string { return d.file("layers.lock") }

// BuilderLog is where the console of the builder VM is kept.
func (d Dir) BuilderLog() string { return d.file("builder.log") }

// CheckLog is where the console of a checked output is kept, for when its
// check fails.
func (d Dir) CheckLog() string { return d.file("check.log") }

func (d Dir) file(name string) string { return filepath.Join(string(d), name) }
