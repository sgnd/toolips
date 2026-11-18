package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// catalog is the stable command table for `toolips list` and help.
var catalog = []catalogEntry{
	{Noun: "pdf", Verbs: []string{"merge", "split", "compress", "rotate", "watermark", "to-image", "from-image", "info", "protect", "unlock", "pages", "number"}, Desc: "Merge, split, compress and convert PDFs"},
	{Noun: "image", Verbs: []string{"compress", "resize", "convert", "crop", "rotate", "watermark", "grayscale", "strip"}, Desc: "Compress, resize, convert and edit images"},
	{Noun: "tls", Verbs: []string{"check"}, Desc: "Check a host's TLS certificate (network)"},
	{Noun: "headers", Verbs: []string{"check"}, Desc: "Grade HTTP security headers (network)"},
	{Noun: "secret", Verbs: []string{"scan"}, Desc: "Scan files for keys and tokens"},
	{Noun: "ct", Verbs: []string{"search"}, Desc: "Search certificate transparency logs (network)"},
	{Noun: "dns", Verbs: []string{"check"}, Desc: "Check a domain's DNS and mail records (network)"},
	{Noun: "hash", Verbs: []string{"sum", "verify"}, Desc: "Checksum files and check them against a list"},
	{Noun: "jwt", Verbs: []string{"decode"}, Desc: "Decode a JSON Web Token locally"},
	{Noun: "version", Verbs: nil, Desc: "Print version"},
	{Noun: "list", Verbs: nil, Desc: "List commands"},
	{Noun: "guide", Verbs: nil, Desc: "Short how-to for people and agents"},
	{Noun: "completion", Verbs: []string{"bash", "zsh", "fish", "powershell"}, Desc: "Shell completion script"},
}

type catalogEntry struct {
	Noun  string
	Verbs []string
	Desc  string
}

func cmdList() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List commands",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			items := make([]listItem, 0, len(catalog))
			for _, e := range catalog {
				items = append(items, listItem{Command: e.Noun, Verbs: e.Verbs, Description: e.Desc})
			}
			return done(cmd, listOut{Commands: items})
		},
	}
}

type listItem struct {
	Command     string   `json:"command"`
	Verbs       []string `json:"verbs,omitempty"`
	Description string   `json:"description"`
}

type listOut struct {
	Commands []listItem `json:"commands"`
}

func (l listOut) Text() string {
	var b strings.Builder
	for _, c := range l.Commands {
		if len(c.Verbs) == 0 {
			fmt.Fprintf(&b, "%-10s  %s\n", c.Command, c.Description)
			continue
		}
		fmt.Fprintf(&b, "%-10s  %s (%s)\n", c.Command, c.Description, strings.Join(c.Verbs, ", "))
	}
	return b.String()
}
