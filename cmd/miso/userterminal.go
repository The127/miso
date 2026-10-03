package main

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"

	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

// userTerminal is what the shell on the other side is told of the terminal
// of the user. It is empty when the user has none.
type userTerminal struct {
	kind       string
	rows, cols uint16
}

// onUserTerminal is the terminal of the user for a shell and what it is
// like. When in is a terminal, it is raw from the moment the shell starts
// until the returned func puts it back, and the sizes it changes to are
// passed on. Anything else is only read.
func onUserTerminal(in io.Reader) (protocol.Terminal, userTerminal, func(), error) {
	file, isFile := in.(*os.File)
	if !isFile || !term.IsTerminal(int(file.Fd())) {
		return protocol.Terminal{In: in}, userTerminal{}, func() {}, nil
	}

	fd := int(file.Fd())
	cols, rows, err := term.GetSize(fd)
	if err != nil {
		return protocol.Terminal{}, userTerminal{}, nil, err
	}

	typed := &rawOnFirstRead{file: file}
	resized, stopFollowing := followSize(fd)
	stop := func() {
		stopFollowing()
		typed.restore()
	}

	like := userTerminal{kind: os.Getenv("TERM"), rows: uint16(rows), cols: uint16(cols)} //nolint:gosec // a terminal is far smaller than 65536 rows or columns

	return protocol.Terminal{In: typed, Resized: resized}, like, stop, nil
}

// followSize answers the sizes the terminal changes to, and a func that ends
// the watching.
func followSize(fd int) (<-chan protocol.Resize, func()) {
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

	return resized, func() {
		signal.Stop(changes)
		close(over)
	}
}

// told is the requests with the shell at their end told what the terminal
// of the user is like.
func told(requests []build.Request, like userTerminal) []build.Request {
	return onShell(requests, func(shell *protocol.Shell) {
		shell.Term, shell.Rows, shell.Cols = like.kind, like.rows, like.cols
	})
}
