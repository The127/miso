package qemu_test

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

// valueOf is what follows the flag in the arguments.
func valueOf(t *testing.T, args []string, flag string) string {
	t.Helper()

	at := slices.Index(args, flag)
	require.NotEqual(t, -1, at, "no %s in %q", flag, args)
	require.Less(t, at+1, len(args), "%s has no value in %q", flag, args)

	return args[at+1]
}

// valuesOf are the values of every flag of that name in args, in order.
func valuesOf(args []string, flag string) []string {
	var values []string
	for at, arg := range args {
		if arg == flag && at+1 < len(args) {
			values = append(values, args[at+1])
		}
	}

	return values
}
