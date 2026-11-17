package pdfx_test

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/sgnd/toolips/internal/pdfx"
)

func TestMergeSplitCompressRotate(t *testing.T) {
	dir := t.TempDir()
	img := filepath.Join(dir, "a.png")
	writePNG(t, img, 20, 20)

	pdf1 := filepath.Join(dir, "one.pdf")
	pdf2 := filepath.Join(dir, "two.pdf")
	ctx := context.Background()
	if err := pdfx.FromImage(ctx, []string{img}, pdf1); err != nil {
		t.Fatal(err)
	}
	if err := pdfx.FromImage(ctx, []string{img}, pdf2); err != nil {
		t.Fatal(err)
	}

	merged := filepath.Join(dir, "merged.pdf")
	if err := pdfx.Merge(ctx, []string{pdf1, pdf2}, merged); err != nil {
		t.Fatal(err)
	}
	if st, err := os.Stat(merged); err != nil || st.Size() == 0 {
		t.Fatalf("merged: %v %#v", err, st)
	}

	splitDir := filepath.Join(dir, "pages")
	if err := pdfx.Split(ctx, merged, splitDir, 1); err != nil {
		t.Fatal(err)
	}
	files, err := pdfx.ListDirFiles(splitDir)
	if err != nil || len(files) < 2 {
		t.Fatalf("split files=%v err=%v", files, err)
	}

	compressed := filepath.Join(dir, "c.pdf")
	if err := pdfx.Compress(ctx, merged, compressed); err != nil {
		t.Fatal(err)
	}

	rotated := filepath.Join(dir, "r.pdf")
	if err := pdfx.Rotate(ctx, pdf1, rotated, 90, nil); err != nil {
		t.Fatal(err)
	}

	wm := filepath.Join(dir, "wm.pdf")
	if err := pdfx.Watermark(ctx, pdf1, wm, "DRAFT", "", nil); err != nil {
		t.Fatal(err)
	}

	imgDir := filepath.Join(dir, "imgs")
	if err := pdfx.ToImage(ctx, pdf1, imgDir, nil); err != nil {
		t.Fatal(err)
	}
}

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{uint8(x * 10), uint8(y * 10), 100, 255})
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
