package main

import (
	"os"
	"sync"

	"golang.org/x/term"
)

// rawOnFirstRead is a terminal that goes raw once it is first read, which is
// when the shell starts. The steps before it print as they always do, and
// Ctrl-C still stops them.
type rawOnFirstRead struct {
	file   *os.File
	once   sync.Once
	before *term.State
	err    error
}

func (r *rawOnFirstRead) Read(p []byte) (int, error) {
	r.once.Do(func() { r.before, r.err = term.MakeRaw(int(r.file.Fd())) })
	if r.err != nil {
		return 0, r.err
	}

	return r.file.Read(p)
}

// restore puts the terminal back, and keeps it from going raw afterwards.
func (r *rawOnFirstRead) restore() {
	r.once.Do(func() {})

	if r.before != nil {
		_ = term.Restore(int(r.file.Fd()), r.before)
	}
}
