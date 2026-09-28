package vercmp

import (
	"cmp"
	"strings"
)

// Compare is negative when version a is older than b, zero when both are
// the same version and positive when a is newer.
func Compare(a, b string) int {
	for a != "" && b != "" {
		if a[0] == '.' && b[0] == '.' {
			a, b = a[1:], b[1:]

			continue
		}

		var numberA, numberB string
		numberA, a = number(a)
		numberB, b = number(b)
		if compared := compareNumbers(numberA, numberB); compared != 0 {
			return compared
		}
	}

	return cmp.Compare(len(a), len(b))
}

// number splits the digits off the front of a version, without the zeros
// they start with.
func number(version string) (digits, rest string) {
	end := 0
	for end < len(version) && '0' <= version[end] && version[end] <= '9' {
		end++
	}

	return strings.TrimLeft(version[:end], "0"), version[end:]
}

func compareNumbers(a, b string) int {
	// a longer number is a bigger one, however many digits it has
	if len(a) != len(b) {
		return cmp.Compare(len(a), len(b))
	}

	return strings.Compare(a, b)
}
