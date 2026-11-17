package imgx_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/sgnd/toolips/internal/imgx"
)

func TestResizeRotateCropGrayscale(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	writePNG(t, src, 40, 20)

	img, format, err := imgx.Open(src)
	if err != nil || format != imgx.FormatPNG {
		t.Fatalf("open: %v %v", err, format)
	}

	resized := imgx.Resize(img, 20, 0)
	if resized.Bounds().Dx() != 20 || resized.Bounds().Dy() != 10 {
		t.Fatalf("resize got %v", resized.Bounds())
	}

	rot, err := imgx.Rotate(img, 90)
	if err != nil {
		t.Fatal(err)
	}
	if rot.Bounds().Dx() != 20 || rot.Bounds().Dy() != 40 {
		t.Fatalf("rotate got %v", rot.Bounds())
	}

	cropped, err := imgx.Crop(img, 5, 5, 10, 8)
	if err != nil {
		t.Fatal(err)
	}
	if cropped.Bounds().Dx() != 10 || cropped.Bounds().Dy() != 8 {
		t.Fatalf("crop got %v", cropped.Bounds())
	}

	gray := imgx.Grayscale(img)
	out := filepath.Join(dir, "gray.png")
	if err := imgx.Save(out, gray, imgx.FormatPNG, 0); err != nil {
		t.Fatal(err)
	}
	wm := imgx.Watermark(img, "toolips")
	jpg := filepath.Join(dir, "out.jpg")
	if err := imgx.Save(jpg, wm, imgx.FormatJPEG, 70); err != nil {
		t.Fatal(err)
	}
}

func TestParseFormat(t *testing.T) {
	if _, err := imgx.ParseFormat("webp"); err == nil {
		t.Fatal("expected webp error")
	}
	f, err := imgx.ParseFormat("jpeg")
	if err != nil || f != imgx.FormatJPEG {
		t.Fatalf("got %v %v", f, err)
	}
}

func TestConvertGIFCompress(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.png")
	writePNG(t, src, 12, 12)
	img, _, err := imgx.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	gifPath := filepath.Join(dir, "out.gif")
	if err := imgx.Save(gifPath, img, imgx.FormatGIF, 0); err != nil {
		t.Fatal(err)
	}
	got, format, err := imgx.Open(gifPath)
	if err != nil || format != imgx.FormatGIF {
		t.Fatalf("gif open: %v %v", err, format)
	}
	jpg := filepath.Join(dir, "out.jpg")
	if err := imgx.Save(jpg, imgx.Compress(got, 60), imgx.FormatJPEG, 60); err != nil {
		t.Fatal(err)
	}
}

func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{200, 100, 50, 255})
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
