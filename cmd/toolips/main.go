// Command toolips is a CLI for everyday PDF, image and security checks,
// for people and AI agents.
package main

import (
	"os"

	"github.com/sgnd/toolips/internal/cli"
)

func main() { os.Exit(cli.Execute()) }
