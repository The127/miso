package main

import (
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"golang.org/x/term"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

// onUserTerminal is the terminal of the user for a shell, and the requests
// with the shell told its type and size. When in is a terminal, it is raw
// from the moment the shell starts until the returned func puts it back, and
// the sizes it changes to are passed on. Anything else is only read.
func onUserTerminal(in io.Reader, requests []build.Request) (protocol.Terminal, []build.Request, func(), error) {
	file, isFile := in.(*os.File)
	if !isFile || !term.IsTerminal(int(file.Fd())) {
		return protocol.Terminal{In: in}, requests, func() {}, nil
	}

	fd := int(file.Fd())
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		return protocol.Terminal{}, nil, nil, err
	}

	typed := &rawOnFirstRead{file: file}

	resized := make(chan protocol.Resize, 1)
	changes := make(chan os.Signal, 1)
	signal.Notify(changes, syscall.SIGWINCH)

	over := make(chan struct{})
	go func() {
		for {
			select {
			case <-changes:
				cols, rows, err := term.GetSize(fd)
				if err != nil {
					continue
				}

				protocol.Latest(resized, protocol.Resize{Rows: uint16(rows), Cols: uint16(cols)}) //nolint:gosec // a terminal is far smaller than 65536 rows or columns
			case <-over:
				return
			}
		}
	}()

	stop := func() {
		signal.Stop(changes)
		close(over)
		typed.restore()
	}

	return protocol.Terminal{In: typed, Resized: resized}, told(requests, os.Getenv("TERM"), uint16(rows), uint16(cols)), stop, nil //nolint:gosec // see above
}

// told is the requests with the shell at their end told the type and size of
// the terminal of the user.
func told(requests []build.Request, kind string, rows, cols uint16) []build.Request {
	told := make([]build.Request, len(requests))
	copy(told, requests)

	last := &told[len(told)-1]
	if shell, isShell := last.Message.(protocol.Shell); isShell {
		shell.Term, shell.Rows, shell.Cols = kind, rows, cols
		last.Message = shell
	}

	return told
}

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
