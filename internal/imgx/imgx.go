// Package imgx provides pure-Go image compress, resize, convert, crop,
// rotate, watermark and grayscale. It reads JPG, PNG, GIF, WebP, BMP and
// TIFF, and writes JPG, PNG and GIF.
package imgx

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "golang.org/x/image/bmp"
	xdraw "golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// Format is an output image format.
type Format string

const (
	FormatJPEG Format = "jpg"
	FormatPNG  Format = "png"
	FormatGIF  Format = "gif"
)

// Writable returns f when toolips can write it, else PNG. WebP, BMP and
// TIFF input then gets a lossless output by default.
func Writable(f Format) Format {
	switch f {
	case FormatJPEG, FormatPNG, FormatGIF:
		return f
	default:
		return FormatPNG
	}
}

// MaxPixels caps width x height before decoding (100 megapixels, about
// 400 MB as RGBA). A small file can declare a huge size and exhaust memory.
const MaxPixels = 100_000_000

// Open decodes an image file.
func Open(path string) (image.Image, Format, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		return nil, "", err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > MaxPixels {
		return nil, "", fmt.Errorf("image is %dx%d, over the %d megapixel limit", cfg.Width, cfg.Height, MaxPixels/1_000_000)
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return nil, "", err
	}
	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", err
	}
	return img, normalizeFormat(format), nil
}

func normalizeFormat(s string) Format {
	switch strings.ToLower(s) {
	case "jpeg", "jpg":
		return FormatJPEG
	case "png":
		return FormatPNG
	case "gif":
		return FormatGIF
	default:
		return Format(strings.ToLower(s))
	}
}

// ParseFormat maps a user string to a Format.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimPrefix(s, ".")) {
	case "jpg", "jpeg":
		return FormatJPEG, nil
	case "png":
		return FormatPNG, nil
	case "gif":
		return FormatGIF, nil
	case "webp":
		return "", fmt.Errorf("webp can be read but not written; use jpg, png or gif")
	default:
		return "", fmt.Errorf("unsupported format %q (use jpg, png or gif)", s)
	}
}

// Save writes img to path in the given format. quality is for JPEG (1–100).
func Save(path string, img image.Image, format Format, quality int) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	switch format {
	case FormatJPEG:
		if quality <= 0 {
			quality = 85
		}
		return jpeg.Encode(f, img, &jpeg.Options{Quality: quality})
	case FormatPNG:
		return png.Encode(f, img)
	case FormatGIF:
		return gif.Encode(f, img, nil)
	default:
		return fmt.Errorf("unsupported format %q", format)
	}
}

// Compress re-encodes with lower quality (JPEG) or as PNG.
func Compress(img image.Image, quality int) image.Image { return img }

// Resize scales to width/height. A zero side keeps aspect ratio.
func Resize(img image.Image, width, height int) image.Image {
	b := img.Bounds()
	ow, oh := b.Dx(), b.Dy()
	if width <= 0 && height <= 0 {
		return img
	}
	if width <= 0 {
		width = ow * height / oh
	}
	if height <= 0 {
		height = oh * width / ow
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), img, b, xdraw.Over, nil)
	return dst
}

// Crop returns the rectangle (x,y,w,h) of img.
func Crop(img image.Image, x, y, w, h int) (image.Image, error) {
	b := img.Bounds()
	r := image.Rect(b.Min.X+x, b.Min.Y+y, b.Min.X+x+w, b.Min.Y+y+h)
	if !r.In(b) {
		return nil, fmt.Errorf("crop rectangle outside image bounds %v", b)
	}
	if sub, ok := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}); ok {
		return sub.SubImage(r), nil
	}
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(dst, dst.Bounds(), img, r.Min, draw.Src)
	return dst, nil
}

// Rotate turns the image clockwise by 90, 180 or 270 degrees.
func Rotate(img image.Image, degrees int) (image.Image, error) {
	degrees = ((degrees % 360) + 360) % 360
	switch degrees {
	case 0:
		return img, nil
	case 90:
		return rotate90(img), nil
	case 180:
		return rotate180(img), nil
	case 270:
		return rotate270(img), nil
	default:
		return nil, fmt.Errorf("rotation must be 90, 180 or 270 (got %d)", degrees)
	}
}

func rotate90(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(h-1-y, x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func rotate180(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(w-1-x, h-1-y, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

func rotate270(img image.Image) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dst.Set(y, w-1-x, img.At(b.Min.X+x, b.Min.Y+y))
		}
	}
	return dst
}

// Grayscale converts to gray.
func Grayscale(img image.Image) image.Image {
	b := img.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			dst.Set(x-b.Min.X, y-b.Min.Y, color.GrayModel.Convert(img.At(x, y)))
		}
	}
	return dst
}

// Watermark draws text in the bottom-right corner.
func Watermark(img image.Image, text string) image.Image {
	b := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(dst, dst.Bounds(), img, b.Min, draw.Src)
	if text == "" {
		return dst
	}
	col := color.RGBA{255, 255, 255, 180}
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(col),
		Face: face,
	}
	adv := d.MeasureString(text)
	pad := 8
	d.Dot = fixed.P(b.Dx()-adv.Ceil()-pad, b.Dy()-pad)
	d.DrawString(text)
	return dst
}

// DefaultOut builds an output path beside inPath.
func DefaultOut(inPath, suffix string, format Format) string {
	dir := filepath.Dir(inPath)
	base := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
	ext := string(format)
	if format == FormatJPEG {
		ext = "jpg"
	}
	return filepath.Join(dir, base+suffix+"."+ext)
}

// ExtFormat guesses format from a path extension.
func ExtFormat(path string) (Format, error) {
	return ParseFormat(filepath.Ext(path))
}
