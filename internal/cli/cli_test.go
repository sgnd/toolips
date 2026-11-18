package cli

import (
	"bytes"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sgnd/toolips/internal/out"
)

func TestExecuteVersionList(t *testing.T) {
	Version = "test"
	code := runCaptured(t, []string{"version", "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	code = runCaptured(t, []string{"list", "-o", "json"})
	if code != 0 {
		t.Fatalf("list exit %d", code)
	}
}

func TestImageResizeDryRun(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "a.png")
	writePNG(t, src, 30, 30)
	code := runCaptured(t, []string{"image", "resize", src, "--width", "10", "--dry-run", "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
}

func TestImageBatchGlobAndDir(t *testing.T) {
	dir := t.TempDir()
	writePNG(t, filepath.Join(dir, "a.png"), 16, 16)
	writePNG(t, filepath.Join(dir, "b.png"), 16, 16)
	_ = os.WriteFile(filepath.Join(dir, "skip.txt"), []byte("nope"), 0o644)

	outDir := filepath.Join(dir, "out")
	code := runCaptured(t, []string{
		"image", "grayscale", filepath.Join(dir, "*.png"),
		"--outdir", outDir, "-o", "json",
	})
	if code != 0 {
		t.Fatalf("glob exit %d", code)
	}
	ents, err := os.ReadDir(outDir)
	if err != nil || len(ents) != 2 {
		t.Fatalf("glob outdir=%v err=%v", ents, err)
	}

	outDir2 := filepath.Join(dir, "out2")
	code = runCaptured(t, []string{"image", "grayscale", dir, "--outdir", outDir2, "-o", "json"})
	if code != 0 {
		t.Fatalf("dir exit %d", code)
	}
	ents, err = os.ReadDir(outDir2)
	if err != nil || len(ents) != 2 {
		t.Fatalf("dir outdir=%v err=%v", ents, err)
	}
}

func TestImageVerbsWrite(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	writePNG(t, src, 32, 24)

	cases := []struct {
		name string
		args []string
		want string
	}{
		{"compress", []string{"image", "compress", src, "--quality", "60", "--out", filepath.Join(dir, "c.jpg"), "-o", "json"}, filepath.Join(dir, "c.jpg")},
		{"convert", []string{"image", "convert", src, "--format", "gif", "--out", filepath.Join(dir, "c.gif"), "-o", "json"}, filepath.Join(dir, "c.gif")},
		{"crop", []string{"image", "crop", src, "--x", "2", "--y", "2", "--width", "10", "--height", "8", "--out", filepath.Join(dir, "crop.png"), "-o", "json"}, filepath.Join(dir, "crop.png")},
		{"rotate", []string{"image", "rotate", src, "--degrees", "180", "--out", filepath.Join(dir, "rot.png"), "-o", "json"}, filepath.Join(dir, "rot.png")},
		{"watermark", []string{"image", "watermark", src, "--text", "toolips", "--out", filepath.Join(dir, "wm.png"), "-o", "json"}, filepath.Join(dir, "wm.png")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code := runCaptured(t, tc.args)
			if code != 0 {
				t.Fatalf("exit %d", code)
			}
			if _, err := os.Stat(tc.want); err != nil {
				t.Fatalf("missing output: %v", err)
			}
		})
	}
}

func TestSecretScanCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.env")
	_ = os.WriteFile(path, []byte("KEY=AKIAIOSFODNN7EXAMPLE\n"), 0o644)
	code := runCaptured(t, []string{"secret", "scan", dir, "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
}

func TestPDFCLIPack(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	writePNG(t, img, 24, 24)

	pdf1 := filepath.Join(dir, "one.pdf")
	pdf2 := filepath.Join(dir, "two.pdf")
	code := runCaptured(t, []string{"pdf", "from-image", img, "--out", pdf1, "-o", "json"})
	if code != 0 {
		t.Fatalf("from-image exit %d", code)
	}
	code = runCaptured(t, []string{"pdf", "to-pdf", img, "--out", pdf2, "-o", "json"})
	if code != 0 {
		t.Fatalf("to-pdf alias exit %d", code)
	}

	merged := filepath.Join(dir, "merged.pdf")
	code = runCaptured(t, []string{"pdf", "merge", pdf1, pdf2, "--out", merged, "-o", "json"})
	if code != 0 {
		t.Fatalf("merge exit %d", code)
	}
	if st, err := os.Stat(merged); err != nil || st.Size() == 0 {
		t.Fatalf("merged missing: %v", err)
	}

	code = runCaptured(t, []string{"pdf", "merge", pdf1, pdf2, "--out", filepath.Join(dir, "dry.pdf"), "--dry-run", "-o", "json"})
	if code != 0 {
		t.Fatalf("merge dry-run exit %d", code)
	}

	splitDir := filepath.Join(dir, "pages")
	code = runCaptured(t, []string{"pdf", "split", merged, "--outdir", splitDir, "--span", "1", "-o", "json"})
	if code != 0 {
		t.Fatalf("split exit %d", code)
	}
	ents, err := os.ReadDir(splitDir)
	if err != nil || len(ents) < 2 {
		t.Fatalf("split out=%v err=%v", ents, err)
	}

	compressed := filepath.Join(dir, "c.pdf")
	code = runCaptured(t, []string{"pdf", "compress", merged, "--out", compressed, "-o", "json"})
	if code != 0 {
		t.Fatalf("compress exit %d", code)
	}

	rotated := filepath.Join(dir, "r.pdf")
	code = runCaptured(t, []string{"pdf", "rotate", pdf1, "--degrees", "90", "--out", rotated, "-o", "json"})
	if code != 0 {
		t.Fatalf("rotate exit %d", code)
	}

	wm := filepath.Join(dir, "wm.pdf")
	code = runCaptured(t, []string{"pdf", "watermark", pdf1, "--text", "DRAFT", "--out", wm, "-o", "json"})
	if code != 0 {
		t.Fatalf("watermark exit %d", code)
	}

	code = runCaptured(t, []string{"pdf", "watermark", pdf1, "-o", "json"})
	if code != out.ExitUsage {
		t.Fatalf("watermark without --text want exit %d got %d", out.ExitUsage, code)
	}

	imgDir := filepath.Join(dir, "imgs")
	code = runCaptured(t, []string{"pdf", "to-image", pdf1, "--outdir", imgDir, "-o", "json"})
	if code != 0 {
		t.Fatalf("to-image exit %d", code)
	}
}

func runCaptured(t *testing.T, args []string) int {
	t.Helper()
	var buf bytes.Buffer
	old := out.P
	out.P = &out.Printer{Format: "json", Version: Version, Start: time.Now(), W: &buf}
	defer func() { out.P = old }()
	code := executeArgs(args)
	if code == 0 && buf.Len() > 0 {
		var env out.Envelope
		if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
			t.Fatalf("json: %v body=%s", err, buf.String())
		}
		if !env.OK {
			t.Fatalf("envelope not ok: %s", buf.String())
		}
	}
	return code
}

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{10, 20, 30, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

func TestMissingInputIsNotFound(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope.pdf")
	cases := [][]string{
		{"pdf", "merge", missing, missing, "-o", "json"},
		{"pdf", "split", missing, "--dry-run", "-o", "json"},
		{"pdf", "compress", missing, "-o", "json"},
		{"pdf", "from-image", missing, "-o", "json"},
		{"secret", "scan", missing, "-o", "json"},
	}
	for _, args := range cases {
		if code := runCaptured(t, args); code != out.ExitNotFound {
			t.Errorf("%v: exit %d, want %d", args, code, out.ExitNotFound)
		}
	}
}
