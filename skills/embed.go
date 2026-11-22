// Package skills holds the agent skill that ships inside the binary.
package skills

import _ "embed"

// Toolips is skills/toolips/SKILL.md. `toolips guide install` writes it.
//
//go:embed toolips/SKILL.md
var Toolips string
