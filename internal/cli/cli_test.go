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

func TestSecretScanCLI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "x.env")
	_ = os.WriteFile(path, []byte("KEY=AKIAIOSFODNN7EXAMPLE\n"), 0o644)
	code := runCaptured(t, []string{"secret", "scan", dir, "-o", "json"})
	if code != 0 {
		t.Fatalf("exit %d", code)
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
