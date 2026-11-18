package out

import (
	"fmt"
	"strings"
)

// SafeText escapes terminal-control characters in untrusted strings (file
// names, header values, CT fields, scanned lines) before they reach a
// terminal. Only tab and newline pass through: a carriage return could
// overwrite a line already printed.
func SafeText(s string) string {
	if !needsSafe(s) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isControl(r) {
			fmt.Fprintf(&b, `\x%02x`, r)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func needsSafe(s string) bool {
	for _, r := range s {
		if isControl(r) {
			return true
		}
	}
	return false
}

func isControl(r rune) bool {
	return (r < 0x20 && r != '\t' && r != '\n') || r == 0x7f || (r >= 0x80 && r <= 0x9f)
}
