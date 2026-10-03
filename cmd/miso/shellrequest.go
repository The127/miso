package main

import (
	"github.com/The127/miso/internal/build"
	"github.com/The127/miso/internal/protocol"
)

// onShell is the requests with the shell at their end changed by edit.
func onShell(requests []build.Request, edit func(shell *protocol.Shell)) []build.Request {
	changed := make([]build.Request, len(requests))
	copy(changed, requests)

	last := &changed[len(changed)-1]
	if shell, isShell := last.Message.(protocol.Shell); isShell {
		edit(&shell)
		last.Message = shell
	}

	return changed
}
