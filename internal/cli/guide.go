package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

const guideText = `toolips — all-in-one everyday tools for people and agents

Grammar: toolips <noun> <verb> [files...] [flags]

Output:
  -o text|json|auto   text for terminals, JSON for agents (default: auto)
  --dry-run           on write commands: show what would be written

First five minutes:
  toolips version
  toolips list
  toolips pdf merge a.pdf b.pdf --out out.pdf
  toolips image resize photo.jpg --width 800 --out small.jpg
  toolips image grayscale ./photos --outdir ./gray
  toolips image convert "*.png" --format jpg --outdir ./jpg
  toolips tls check example.com

Image notes:
  Reads JPG, PNG, GIF, WebP, BMP and TIFF. Writes JPG, PNG and GIF;
  WebP, BMP and TIFF input is written as PNG unless --out or --format
  names another format.

PDF notes:
  pdf to-image extracts embedded images (pure Go). True page
  rasterization needs CGO and is not included.

Agents: always pass -o json (or pipe stdout). Read error.hint on failure.
Exit codes: 0 ok, 1 error, 2 usage, 3 not found, 4 network.
`

func cmdGuide() *cobra.Command {
	return &cobra.Command{
		Use:   "guide",
		Short: "Short how-to for people and agents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return done(cmd, guideOut{Guide: strings.TrimSpace(guideText)})
		},
	}
}

type guideOut struct {
	Guide string `json:"guide"`
}

func (g guideOut) Text() string { return g.Guide + "\n" }
