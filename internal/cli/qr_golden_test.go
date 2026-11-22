package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"rsc.io/qr"
)

// The matrices in testdata/qr were written once from rsc.io/qr v0.2.0 and
// then checked by eye (finder squares, timing lines). They are not from an
// independent encoder: they catch a library change or an update that alters
// the code for the same text. Set UPDATE_QR=1 to rewrite them on purpose.
var qrGolden = []struct {
	file, text string
	level      qr.Level
}{
	{"hello-L.txt", "HELLO", qr.L},
	{"url-M.txt", "https://toolips.efolusi.com", qr.M},
	{"digits-H.txt", "0123456789012345", qr.H},
}

func matrixText(c *qr.Code) string {
	var b strings.Builder
	for y := 0; y < c.Size; y++ {
		for x := 0; x < c.Size; x++ {
			if c.Black(x, y) {
				b.WriteByte('#')
			} else {
				b.WriteByte('.')
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func TestQRMatrixGolden(t *testing.T) {
	for _, g := range qrGolden {
		c, err := qr.Encode(g.text, g.level)
		if err != nil {
			t.Fatal(err)
		}
		got := matrixText(c)
		path := filepath.Join("testdata", "qr", g.file)
		if os.Getenv("UPDATE_QR") == "1" {
			if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if got != string(want) {
			t.Errorf("%s: the code for %q changed", g.file, g.text)
		}
		// timing patterns: row 6 and column 6 alternate between the finders
		for i := 8; i < c.Size-8; i++ {
			if c.Black(i, 6) != (i%2 == 0) || c.Black(6, i) != (i%2 == 0) {
				t.Errorf("%s: timing pattern broken at %d", g.file, i)
				break
			}
		}
	}
}
