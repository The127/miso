// Package sandbox runs the command of a run in a root of its own, with its
// own processes, mounts, name and network, and nothing of the builder above
// it. The agent starts itself again as the helper that sets this up and
// then becomes the shell.
package sandbox
