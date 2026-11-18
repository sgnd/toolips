package imgx_test

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"bytes"
	"encoding/binary"
	"github.com/sgnd/toolips/internal/imgx"
	"hash/crc32"
	"strings"
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

func TestRotateCropErrors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	if _, err := imgx.Rotate(img, 45); err == nil {
		t.Fatal("expected rotate error")
	}
	if _, err := imgx.Crop(img, 0, 0, 20, 20); err == nil {
		t.Fatal("expected crop bounds error")
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

func TestParseAndExtFormat(t *testing.T) {
	ok := map[string]imgx.Format{"jpg": imgx.FormatJPEG, ".JPEG": imgx.FormatJPEG, "png": imgx.FormatPNG, "gif": imgx.FormatGIF}
	for in, want := range ok {
		got, err := imgx.ParseFormat(in)
		if err != nil || got != want {
			t.Errorf("ParseFormat(%q) = %q, %v", in, got, err)
		}
	}
	for _, in := range []string{"webp", "bmp", ""} {
		if _, err := imgx.ParseFormat(in); err == nil {
			t.Errorf("ParseFormat(%q) accepted", in)
		}
	}
	if f, err := imgx.ExtFormat("/a/b.PNG"); err != nil || f != imgx.FormatPNG {
		t.Errorf("ExtFormat = %q, %v", f, err)
	}
}

func TestDefaultOut(t *testing.T) {
	got := imgx.DefaultOut(filepath.Join("dir", "a.png"), "_small", imgx.FormatJPEG)
	if want := filepath.Join("dir", "a_small.jpg"); got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestRotateAndCropBounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for deg, want := range map[int]image.Point{0: {4, 2}, 90: {2, 4}, 180: {4, 2}, 270: {2, 4}, -90: {2, 4}} {
		r, err := imgx.Rotate(img, deg)
		if err != nil || r.Bounds().Size() != want {
			t.Errorf("Rotate %d = %v, %v; want %v", deg, r.Bounds().Size(), err, want)
		}
	}
	if _, err := imgx.Rotate(img, 45); err == nil {
		t.Error("Rotate 45 accepted")
	}
	if _, err := imgx.Crop(img, 2, 0, 4, 2); err == nil {
		t.Error("crop outside bounds accepted")
	}
}

// TestOpenRejectsHugeDeclaredSize writes a tiny PNG whose header claims
// 20000x20000 pixels; Open must refuse it before allocating.
func TestOpenRejectsHugeDeclaredSize(t *testing.T) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewGray(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	b := buf.Bytes()
	// IHDR data starts at byte 16: width, height (big endian), then CRC at 29.
	binary.BigEndian.PutUint32(b[16:], 20000)
	binary.BigEndian.PutUint32(b[20:], 20000)
	binary.BigEndian.PutUint32(b[29:], crc32.ChecksumIEEE(b[12:29]))
	path := filepath.Join(t.TempDir(), "bomb.png")
	if err := os.WriteFile(path, b, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, err := imgx.Open(path); err == nil || !strings.Contains(err.Error(), "megapixel") {
		t.Fatalf("got %v, want megapixel limit error", err)
	}
}
