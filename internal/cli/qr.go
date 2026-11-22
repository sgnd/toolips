package cli

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"rsc.io/qr"

	"github.com/sgnd/toolips/internal/out"
)

const (
	qrQuietZone = 4    // modules of white around the code, as the QR standard asks
	qrMaxScale  = 100  // pixels per module
	qrMaxSide   = 8192 // pixels on each side of the image; about 67 MB in memory
	qrMaxInput  = 4096 // bytes read from stdin; the largest code holds 2953
)

var qrLevels = map[string]qr.Level{"L": qr.L, "M": qr.M, "Q": qr.Q, "H": qr.H}

func cmdQR() *cobra.Command {
	q := &cobra.Command{Use: "qr", Short: "QR codes"}
	q.AddCommand(qrMake())
	return q
}

func qrMake() *cobra.Command {
	var level string
	var scale int
	cmd := &cobra.Command{
		Use:   "make TEXT",
		Short: "Write a QR code for a text or URL as a PNG",
		Long: `Write a QR code for TEXT as a black and white PNG.

TEXT is the text or URL the code holds. Pass "-" to read it from stdin (first
line). --level is the error correction: L, M (default), Q or H; a higher level
survives more damage and makes a bigger code. --scale is the pixels per
module (default 8); the image may be at most 8192 pixels on a side, so a
large code needs a smaller scale. The code has the standard 4-module white border.
The image goes to --out (default qr.png); an existing file stops the command
unless --force is set. Everything is made on this machine.`,
		Example: `  toolips qr make https://example.com --out site.png
  echo "WIFI:T:WPA;S:home;P:secret;;" | toolips qr make - --out wifi.png`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			lv, ok := qrLevels[strings.ToUpper(level)]
			if !ok {
				return out.Usage("--level must be L, M, Q or H")
			}
			if scale < 1 || scale > qrMaxScale {
				return out.Usage("--scale must be between 1 and %d", qrMaxScale)
			}
			text := args[0]
			if text == "-" {
				var err error
				if text, err = readSecretLine(cmd, "-", "text", "pipe the text on stdin", qrMaxInput); err != nil {
					return err
				}
			}
			if text == "" {
				return out.Usage("the text is empty")
			}
			outPath, _ := writeFlag(cmd)
			if outPath == "" {
				outPath = "qr.png"
			}
			if err := refuseOverwrite(cmd, outPath); err != nil {
				return err
			}
			code, err := qr.Encode(text, lv)
			if err != nil {
				return out.Errf(out.ExitError, "QR_TOO_LONG", "shorten the text, or use --level L", "the text does not fit in a QR code at level %s", strings.ToUpper(level))
			}
			if side := (code.Size + 2*qrQuietZone) * scale; side > qrMaxSide {
				return out.Usage("--scale %d makes an image %d pixels wide; the limit is %d, so use --scale %d or less", scale, side, qrMaxSide, qrMaxSide/(code.Size+2*qrQuietZone))
			}
			if g.dryRun {
				return done(cmd, newFilesOut([]string{outPath}, "would write "+outPath))
			}
			if err := os.WriteFile(outPath, renderQR(code, scale), 0o644); err != nil {
				return out.Errf(out.ExitError, "WRITE_FAILED", "check that you may write to this folder", "write %s: %v", outPath, err)
			}
			return done(cmd, newFilesOut([]string{outPath}, ""))
		},
	}
	cmd.Flags().StringVar(&level, "level", "M", "error correction: L, M, Q or H")
	cmd.Flags().IntVar(&scale, "scale", 8, "pixels per module (image side at most 8192)")
	cmd.Flags().String("out", "", "output file path (default qr.png)")
	cmd.Flags().Bool("force", false, "replace the output file if it exists")
	return cmd
}

// renderQR draws code as a PNG: black modules on white, with the quiet zone.
func renderQR(code *qr.Code, scale int) []byte {
	n := (code.Size + 2*qrQuietZone) * scale
	img := image.NewGray(image.Rect(0, 0, n, n))
	for i := range img.Pix {
		img.Pix[i] = 0xFF
	}
	for my := 0; my < code.Size; my++ {
		for mx := 0; mx < code.Size; mx++ {
			if !code.Black(mx, my) {
				continue
			}
			x0, y0 := (mx+qrQuietZone)*scale, (my+qrQuietZone)*scale
			for y := y0; y < y0+scale; y++ {
				for x := x0; x < x0+scale; x++ {
					img.SetGray(x, y, color.Gray{Y: 0})
				}
			}
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img) // encoding to memory cannot fail for a Gray image
	return b.Bytes()
}
