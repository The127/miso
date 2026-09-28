package vercmp

import (
	"cmp"
	"strings"
)

// Compare is negative when version a is older than b, zero when both are
// the same version and positive when a is newer.
func Compare(a, b string) int {
	for {
		a, b = skipped(a), skipped(b)
		compared, found := separator(a, b, tilde)
		if !found {
			if a == "" || b == "" {
				return cmp.Compare(len(a), len(b))
			}

			compared, found = separator(a, b, separators)
		}

		if found {
			if compared != 0 {
				return compared
			}

			a, b = a[1:], b[1:]

			continue
		}

		if isDigit(a[0]) || isDigit(b[0]) {
			var numberA, numberB string
			numberA, a = number(a)
			numberB, b = number(b)
			if compared := compareNumbers(numberA, numberB); compared != 0 {
				return compared
			}

			continue
		}

		var lettersA, lettersB string
		lettersA, a = letters(a)
		lettersB, b = letters(b)
		// byte order puts capitals first, as the specification wants
		if compared := strings.Compare(lettersA, lettersB); compared != 0 {
			return compared
		}
	}
}

// tilde is looked for before a version that has ended counts as older, so
// it makes the older version even against the end.
const tilde = "~"

// separators are looked for in the order the specification checks them.
const separators = "-^."

// separator compares the fronts of two versions when either starts with one
// of the separators among. The version with a separator the other lacks is
// the older one.
func separator(a, b, among string) (compared int, found bool) {
	for _, each := range among {
		atA, atB := strings.HasPrefix(a, string(each)), strings.HasPrefix(b, string(each))
		switch {
		case atA && atB:
			return 0, true
		case atA:
			return -1, true
		case atB:
			return 1, true
		}
	}

	return 0, false
}

// skipped is a version without the characters at its front that the
// format has no meaning for.
func skipped(version string) string {
	// byte by byte, every byte of a character beyond ASCII is skipped
	start := 0
	for start < len(version) && !isDigit(version[start]) && !isLetter(version[start]) && strings.IndexByte("-.~^", version[start]) < 0 {
		start++
	}

	return version[start:]
}

// number splits the digits off the front of a version, without the zeros
// they start with.
func number(version string) (digits, rest string) {
	end := 0
	for end < len(version) && isDigit(version[end]) {
		end++
	}

	return strings.TrimLeft(version[:end], "0"), version[end:]
}

// letters splits the letters off the front of a version.
func letters(version string) (front, rest string) {
	end := 0
	for end < len(version) && isLetter(version[end]) {
		end++
	}

	return version[:end], version[end:]
}

func isDigit(c byte) bool {
	return '0' <= c && c <= '9'
}

func isLetter(c byte) bool {
	return 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z'
}

func compareNumbers(a, b string) int {
	// a longer number is a bigger one, however many digits it has
	if len(a) != len(b) {
		return cmp.Compare(len(a), len(b))
	}

	return strings.Compare(a, b)
}
